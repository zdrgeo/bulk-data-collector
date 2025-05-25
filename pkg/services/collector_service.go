package services

import (
	"context"
	"errors"
	"time"
)

var (
	ErrBackpressure = errors.New("backpressure")
)

type ReportModel struct {
	CollectionTime time.Time
	Parameters     map[string]any
}

type DataModel struct {
	Reports []*ReportModel
}

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
	Collect(ctx context.Context, oui, productClass, serialNumber string, data *DataModel) error
	CollectCSV(ctx context.Context, oui, productClass, serialNumber string, bulkData *CSVBulkDataModel) error
	CollectJSON(ctx context.Context, oui, productClass, serialNumber string, bulkData *JSONBulkDataModel) error
}
