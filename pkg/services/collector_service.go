package services

import (
	"context"
	"errors"
	"time"
)

var (
	ErrBackpressure = errors.New("backpressure")
)

type ParameterPerRowModel struct {
	ReportTimestamp time.Time
	ParameterName   string
	ParameterValue  string
	ParameterType   string
}

type CSVBulkDataModel struct {
	ParameterPerRow []*ParameterPerRowModel
}

type NameValuePairModel struct {
	Report []map[string]any `json:"Report"`
}

type JSONBulkDataModel struct {
	NameValuePair *NameValuePairModel
}

type CollectorService interface {
	CollectCSV(ctx context.Context, oui, productClass, serialNumber string, bulkData *CSVBulkDataModel) error
	CollectJSON(ctx context.Context, oui, productClass, serialNumber string, bulkData *JSONBulkDataModel) error
}
