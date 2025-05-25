package services

import (
	"context"
)

type MockCollectorService struct{}

var _ CollectorService = (*MockCollectorService)(nil)

func NewMockCollectorService() *MockCollectorService {
	return &MockCollectorService{}
}

func (s *MockCollectorService) Collect(ctx context.Context, oui, productClass, serialNumber string, data *DataModel) error {
	return nil
}

func (s *MockCollectorService) CollectCSV(ctx context.Context, oui, productClass, serialNumber string, bulkData *CSVBulkDataModel) error {
	return nil
}

func (s *MockCollectorService) CollectJSON(ctx context.Context, oui, productClass, serialNumber string, bulkData *JSONBulkDataModel) error {
	return nil
}
