package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/zdrgeo/bulk-data-collector/pkg/services"
)

type OTelMeterOptions struct {
	Name        string
	Instruments []*OTelInstrumentOptions
}

type OTelCollectorServiceOptions struct {
	Meter *OTelMeterOptions
}

type OTelCollectorService struct {
	instruments map[string]OTelInstrument
	options     *OTelCollectorServiceOptions
}

func NewOTelCollectorService(options *OTelCollectorServiceOptions) (*OTelCollectorService, error) {
	meter := otel.Meter(options.Meter.Name)

	instruments := make(map[string]OTelInstrument, len(options.Meter.Instruments))

	for _, instrumentOptions := range options.Meter.Instruments {
		instrument, err := NewOTelInstrument(meter, instrumentOptions)

		if err != nil {
			return nil, err
		}

		instruments[instrumentOptions.ParameterName] = instrument
	}

	return &OTelCollectorService{instruments: instruments, options: options}, nil
}

func (s *OTelCollectorService) CollectCSV(ctx context.Context, oui, productClass, serialNumber string, bulkData *services.CSVBulkDataModel) error {
	reports := map[time.Time][]*services.ParameterPerRowModel{}

	for _, parameterPerRow := range bulkData.ParameterPerRow {
		reports[parameterPerRow.ReportTimestamp] = append(reports[parameterPerRow.ReportTimestamp], parameterPerRow)
	}

	for _, report := range reports {
		attributes := attribute.NewSet(attribute.String("OUI", oui), attribute.String("ProductClass", productClass), attribute.String("SerialNumber", serialNumber))

		for _, parameterPerRow := range report {
			value, err := services.ParseParameterValue(parameterPerRow.ParameterType, parameterPerRow.ParameterValue)

			if err != nil {
				return err
			}

			instrument, ok := s.instruments[parameterPerRow.ParameterName]

			if ok {
				if err := instrument.Measure(ctx, value, attributes); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *OTelCollectorService) CollectJSON(ctx context.Context, oui, productClass, serialNumber string, bulkData *services.JSONBulkDataModel) error {
	attributes := attribute.NewSet(attribute.String("OUI", oui), attribute.String("ProductClass", productClass), attribute.String("SerialNumber", serialNumber))

	if bulkData.NameValuePair != nil {
		for _, report := range bulkData.NameValuePair.Report {
			for key, value := range report {
				instrument, ok := s.instruments[key]

				if ok {
					if err := instrument.Measure(ctx, value, attributes); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
