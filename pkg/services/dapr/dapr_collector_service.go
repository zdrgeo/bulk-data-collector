package dapr

import (
	"context"
	"fmt"
	"time"

	daprclient "github.com/dapr/go-sdk/client"

	"github.com/zdrgeo/bulk-data-collector/pkg/services"
)

type DaprEventModel struct {
	CollectionTime time.Time
	OUI            string         `json:"OUI"`
	ProductClass   string         `json:"ProductClass"`
	SerialNumber   string         `json:"SerialNumber"`
	Parameters     map[string]any `json:"Parameters"`
}

type DaprCollectorServiceOptions struct {
	PubSubName string
	TopicName  string
}

type DaprCollectorService struct {
	daprClient daprclient.Client
	options    *DaprCollectorServiceOptions
}

var _ services.CollectorService = (*DaprCollectorService)(nil)

func NewDaprCollectorService(daprClient daprclient.Client, option *DaprCollectorServiceOptions) *DaprCollectorService {
	return &DaprCollectorService{daprClient: daprClient, options: option}
}

func (s *DaprCollectorService) CollectCSV(ctx context.Context, oui, productClass, serialNumber string, bulkData *services.CSVBulkDataModel) error {
	reports := map[time.Time][]*services.ParameterPerRowModel{}

	for _, parameterPerRow := range bulkData.ParameterPerRow {
		reports[parameterPerRow.ReportTimestamp] = append(reports[parameterPerRow.ReportTimestamp], parameterPerRow)
	}

	for reportTimestamp, report := range reports {
		event := &DaprEventModel{
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

		deviceName := fmt.Sprintf("%s-%s-%s", oui, productClass, serialNumber)
		topicName := fmt.Sprintf("%s/device/%s/event", s.options.TopicName, deviceName)

		if err := s.daprClient.PublishEvent(ctx, s.options.PubSubName, topicName, event); err != nil {
			return err
		}
	}

	return nil
}

func (s *DaprCollectorService) CollectJSON(ctx context.Context, oui, productClass, serialNumber string, bulkData *services.JSONBulkDataModel) error {
	if bulkData.NameValuePair != nil {
		for _, report := range bulkData.NameValuePair.Report {
			event := &DaprEventModel{
				CollectionTime: report["CollectionTime"].(time.Time),
				OUI:            oui,
				ProductClass:   productClass,
				SerialNumber:   serialNumber,
				Parameters:     make(map[string]any, len(report)),
			}

			for key, value := range report {
				event.Parameters[key] = value
			}

			deviceName := fmt.Sprintf("%s-%s-%s", oui, productClass, serialNumber)
			topicName := fmt.Sprintf("%s/device/%s/event", s.options.TopicName, deviceName)

			if err := s.daprClient.PublishEvent(ctx, s.options.PubSubName, topicName, event); err != nil {
				return err
			}
		}
	}

	return nil
}
