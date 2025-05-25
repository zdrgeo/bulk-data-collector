package azureeventhubs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/zdrgeo/bulk-data-collector/pkg/services"
)

// https://github.com/Azure/azure-sdk-for-net/blob/main/sdk/eventhub/Azure.Messaging.EventHubs/samples/Sample04_PublishingEvents.md#tuning-throughput-for-buffered-publishing

const (
	meterName = "collector"
)

type AzureEventHubsEventModel struct {
	CollectionTime time.Time      `json:"CollectionTime"`
	OUI            string         `json:"OUI"`
	ProductClass   string         `json:"ProductClass"`
	SerialNumber   string         `json:"SerialNumber"`
	Parameters     map[string]any `json:"Parameters"`
}

type AzureEventHubsEventBatchModel struct {
	Events []*AzureEventHubsEventModel
}

type AzureEventHubsCollectorServiceOptions struct {
	PartitionQueueLimit     int
	PartitionProducersCount int
}

type partitionQueue struct {
	partitionID string
	queue       chan *AzureEventHubsEventModel
}

type AzureEventHubsCollectorService struct {
	producerClient  *azeventhubs.ProducerClient
	options         *AzureEventHubsCollectorServiceOptions
	partitionQueues []*partitionQueue
	queueCounter    metric.Int64UpDownCounter
	batchCounter    metric.Int64Counter
	eventCounter    metric.Int64Counter
}

var _ services.CollectorService = (*AzureEventHubsCollectorService)(nil)

func NewAzureEventHubsCollectorService(producerClient *azeventhubs.ProducerClient, options *AzureEventHubsCollectorServiceOptions) (*AzureEventHubsCollectorService, error) {
	meter := otel.Meter(meterName)

	queueCounter, err := meter.Int64UpDownCounter("partition_queue_counter", metric.WithDescription("Partition queue counter"), metric.WithUnit("count"))

	if err != nil {
		return nil, err
	}

	batchCounter, err := meter.Int64Counter("partition_batch_counter", metric.WithDescription("Partition batch counter"), metric.WithUnit("count"))

	if err != nil {
		return nil, err
	}

	eventCounter, err := meter.Int64Counter("partition_event_counter", metric.WithDescription("Partition event counter"), metric.WithUnit("count"))

	if err != nil {
		return nil, err
	}

	partitionQueueLimit := 1_000

	if options != nil {
		if options.PartitionQueueLimit > 0 {
			partitionQueueLimit = options.PartitionQueueLimit
		}
	}

	eventHubProperties, err := producerClient.GetEventHubProperties(context.Background(), nil)

	if err != nil {
		return nil, err
	}

	partitionQueues := make([]*partitionQueue, 0, len(eventHubProperties.PartitionIDs))

	for _, partitionID := range eventHubProperties.PartitionIDs {
		partitionQueue := &partitionQueue{partitionID: partitionID, queue: make(chan *AzureEventHubsEventModel, partitionQueueLimit)}

		partitionQueues = append(partitionQueues, partitionQueue)
	}

	return &AzureEventHubsCollectorService{producerClient: producerClient, options: options, partitionQueues: partitionQueues, queueCounter: queueCounter, batchCounter: batchCounter, eventCounter: eventCounter}, nil
}

func (s *AzureEventHubsCollectorService) Collect(ctx context.Context, oui, productClass, serialNumber string, data *services.DataModel) error {
	for _, report := range data.Reports {
		event := &AzureEventHubsEventModel{
			CollectionTime: report.CollectionTime,
			OUI:            oui,
			ProductClass:   productClass,
			SerialNumber:   serialNumber,
			Parameters:     make(map[string]any, len(report.Parameters)),
		}

		for key, value := range report.Parameters {
			event.Parameters[key] = value
		}

		if err := s.enqueue(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (s *AzureEventHubsCollectorService) CollectCSV(ctx context.Context, oui, productClass, serialNumber string, bulkData *services.CSVBulkDataModel) error {
	reports := map[time.Time][]*services.ParameterPerRowModel{}

	for _, parameterPerRow := range bulkData.ParameterPerRow {
		reports[parameterPerRow.ReportTimestamp] = append(reports[parameterPerRow.ReportTimestamp], parameterPerRow)
	}

	for reportTimestamp, report := range reports {
		event := &AzureEventHubsEventModel{
			CollectionTime: reportTimestamp,
			OUI:            oui,
			ProductClass:   productClass,
			SerialNumber:   serialNumber,
			Parameters:     make(map[string]any, len(report)),
		}

		for _, parameterPerRow := range report {
			value, err := services.ParseParameterValue(parameterPerRow.ParameterType, parameterPerRow.ParameterValue)

			if err != nil {
				return err
			}

			event.Parameters[parameterPerRow.ParameterName] = value
		}

		if err := s.enqueue(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (s *AzureEventHubsCollectorService) CollectJSON(ctx context.Context, oui, productClass, serialNumber string, bulkData *services.JSONBulkDataModel) error {
	if bulkData.NameValuePair != nil {
		for _, report := range bulkData.NameValuePair.Report {
			event := &AzureEventHubsEventModel{
				CollectionTime: report["CollectionTime"].(time.Time),
				OUI:            oui,
				ProductClass:   productClass,
				SerialNumber:   serialNumber,
				Parameters:     make(map[string]any, len(report)),
			}

			for key, value := range report {
				event.Parameters[key] = value
			}

			if err := s.enqueue(ctx, event); err != nil {
				return err
			}
		}
	}

	return nil
}

type RunError struct {
	PartitionProducerErrs []error
}

func (runErr *RunError) Error() string {
	errMsgs := make([]string, len(runErr.PartitionProducerErrs))

	for _, partitionProducerErr := range runErr.PartitionProducerErrs {
		errMsgs = append(errMsgs, partitionProducerErr.Error())
	}

	errMsg := strings.Join(errMsgs, "\n")

	return errMsg
}

func (s *AzureEventHubsCollectorService) Run(ctx context.Context) error {
	partitionProducersCount := 1
	runToCompletion := false

	if s.options != nil {
		if s.options.PartitionProducersCount > 0 {
			partitionProducersCount = s.options.PartitionProducersCount
		}
	}

	var partitonProducerCtx context.Context

	if runToCompletion {
		partitonProducerCtx = context.Background()
	} else {
		partitonProducerCtx = ctx
	}

	partitionProducerErrs := make(chan error, len(s.partitionQueues)*partitionProducersCount)

	partitionProducerGroup := sync.WaitGroup{}

	partitionProducerGroup.Add(len(s.partitionQueues) * partitionProducersCount)

	for _, partitionQueue := range s.partitionQueues {
		for range partitionProducersCount {
			go func() {
				defer partitionProducerGroup.Done()

				partitionProducerErrs <- s.produce(partitonProducerCtx, partitionQueue)
			}()
		}
	}

	partitionProducerGroup.Wait()

	close(partitionProducerErrs)

	errs := make([]error, 0, len(s.partitionQueues)*partitionProducersCount)

	for partitionProducerErr := range partitionProducerErrs {
		if partitionProducerErr != nil {
			errs = append(errs, partitionProducerErr)
		}
	}

	if len(errs) != 0 {
		return &RunError{
			PartitionProducerErrs: errs,
		}
	}

	return nil
}

func (s *AzureEventHubsCollectorService) enqueue(ctx context.Context, event *AzureEventHubsEventModel) error {
	if len(s.partitionQueues) == 0 {
		return nil
	}

	deviceName := fmt.Sprintf("%s-%s-%s", event.OUI, event.ProductClass, event.SerialNumber)

	hash32 := fnv.New32a()

	hash32.Write([]byte(deviceName))

	deviceHash := hash32.Sum32()

	partitionQueueIndex := int(deviceHash % uint32(len(s.partitionQueues)))

	partitionQueue := s.partitionQueues[partitionQueueIndex]

	select {
	case <-ctx.Done():
		return ctx.Err()
	case partitionQueue.queue <- event:
		s.queueCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))
		// default:
		// 	return ErrBackpressure
	}

	return nil
}

func (s *AzureEventHubsCollectorService) produce(ctx context.Context, partitionQueue *partitionQueue) error {
	eventDataBatchOptions := &azeventhubs.EventDataBatchOptions{
		PartitionID: &partitionQueue.partitionID,
	}

	eventDataBatch, err := s.producerClient.NewEventDataBatch(ctx, eventDataBatchOptions)

	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			numEvents := eventDataBatch.NumEvents()

			if numEvents != 0 {
				if err := s.producerClient.SendEventDataBatch(ctx, eventDataBatch, nil); err != nil {
					return err
				}

				s.batchCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))
				s.eventCounter.Add(ctx, int64(numEvents), metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))
			}

			return ctx.Err()

		case <-time.Tick(1 * time.Minute):
			numEvents := eventDataBatch.NumEvents()

			if numEvents != 0 {
				if err := s.producerClient.SendEventDataBatch(ctx, eventDataBatch, nil); err != nil {
					return err
				}

				s.batchCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))
				s.eventCounter.Add(ctx, int64(numEvents), metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))

				newEventDataBatch, err := s.producerClient.NewEventDataBatch(ctx, eventDataBatchOptions)

				if err != nil {
					return err
				}

				eventDataBatch = newEventDataBatch
			}

		case event, ok := <-partitionQueue.queue:
			if !ok {
				numEvents := eventDataBatch.NumEvents()

				if numEvents != 0 {
					if err := s.producerClient.SendEventDataBatch(ctx, eventDataBatch, nil); err != nil {
						return err
					}

					s.batchCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))
					s.eventCounter.Add(ctx, int64(numEvents), metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))
				}

				return nil
			}

			s.queueCounter.Add(ctx, -1, metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))

			body, err := json.MarshalIndent(event, "", "  ")

			if err != nil {
				return err
			}

			eventData := &azeventhubs.EventData{
				Body: body,
			}

			if err := eventDataBatch.AddEventData(eventData, nil); err != nil {
				if !errors.Is(err, azeventhubs.ErrEventDataTooLarge) {
					return err
				}

				numEvents := eventDataBatch.NumEvents()

				if numEvents == 0 {
					return err
				}

				if err := s.producerClient.SendEventDataBatch(ctx, eventDataBatch, nil); err != nil {
					return err
				}

				s.batchCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))
				s.eventCounter.Add(ctx, int64(numEvents), metric.WithAttributes(attribute.String("partition", partitionQueue.partitionID)))

				newEventDataBatch, err := s.producerClient.NewEventDataBatch(ctx, eventDataBatchOptions)

				if err != nil {
					return err
				}

				if err := newEventDataBatch.AddEventData(eventData, nil); err != nil {
					return err
				}

				eventDataBatch = newEventDataBatch
			}
		}
	}
}
