package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"

	"github.com/zdrgeo/bulk-data-collector/pkg/services"
)

type MQTTEventModel struct {
	OUI          string         `json:"OUI"`
	ProductClass string         `json:"ProductClass"`
	SerialNumber string         `json:"SerialNumber"`
	Parameters   map[string]any `json:"Parameters"`
}

type MQTTCollectorServiceOptions struct {
	TopicName string
}

type MQTTCollectorService struct {
	connectionManager *autopaho.ConnectionManager
	options           *MQTTCollectorServiceOptions
}

func NewMQTTCollectorService(connectionManager *autopaho.ConnectionManager, options *MQTTCollectorServiceOptions) *MQTTCollectorService {
	return &MQTTCollectorService{connectionManager: connectionManager, options: options}
}

func (s *MQTTCollectorService) CollectCSV(ctx context.Context, oui, productClass, serialNumber string, bulkData *services.CSVBulkDataModel) error {
	reports := map[time.Time][]*services.ParameterPerRowModel{}

	for _, parameterPerRow := range bulkData.ParameterPerRow {
		reports[parameterPerRow.ReportTimestamp] = append(reports[parameterPerRow.ReportTimestamp], parameterPerRow)
	}

	for _, report := range reports {
		event := &MQTTEventModel{
			OUI:          oui,
			ProductClass: productClass,
			SerialNumber: serialNumber,
			Parameters:   make(map[string]any, len(report)),
		}

		for _, parameterPerRow := range report {
			value, err := services.ParseParameterValue(parameterPerRow.ParameterType, parameterPerRow.ParameterValue)

			if err != nil {
				return err
			}

			event.Parameters[parameterPerRow.ParameterName] = value
		}

		deviceName := fmt.Sprintf("%s-%s-%s", oui, productClass, serialNumber)
		collectorName := "100000"
		topic := fmt.Sprintf("collector/%s/device/%s/event", collectorName, deviceName)

		payload, err := json.MarshalIndent(event, "", "  ")

		if err != nil {
			return err
		}

		publish := &autopaho.QueuePublish{
			Publish: &paho.Publish{
				Topic:   topic,
				QoS:     1,
				Payload: payload,
			},
		}

		if err := s.connectionManager.PublishViaQueue(ctx, publish); err != nil {
			return err
		}
	}

	return nil
}

func (s *MQTTCollectorService) CollectJSON(ctx context.Context, oui, productClass, serialNumber string, bulkData *services.JSONBulkDataModel) error {
	if bulkData.NameValuePair != nil {
		for _, report := range bulkData.NameValuePair.Report {
			event := &MQTTEventModel{
				OUI:          oui,
				ProductClass: productClass,
				SerialNumber: serialNumber,
				Parameters:   make(map[string]any, len(report)),
			}

			for key, value := range report {
				event.Parameters[key] = value
			}

			deviceName := fmt.Sprintf("%s-%s-%s", oui, productClass, serialNumber)
			collectorName := "100000"
			topic := fmt.Sprintf("collector/%s/device/%s/event", collectorName, deviceName)

			payload, err := json.MarshalIndent(event, "", "  ")

			if err != nil {
				return err
			}

			publish := &autopaho.QueuePublish{
				Publish: &paho.Publish{
					Topic:   topic,
					QoS:     1,
					Payload: payload,
				},
			}

			if err := s.connectionManager.PublishViaQueue(ctx, publish); err != nil {
				return err
			}
		}
	}

	return nil
}
