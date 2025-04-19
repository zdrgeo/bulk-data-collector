package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	collectorservices "github.com/zdrgeo/bulk-data-collector/pkg/services"
)

const (
	// TR-069 and TR-369 report formats
	ReportFormat_ParameterPerRow    = "ParameterPerRow"
	ReportFormat_ParameterPerColumn = "ParameterPerColumn"
	ReportFormat_NameValuePair      = "NameValuePair"
	ReportFormat_ObjectHierarchy    = "ObjectHierarchy"
	// TR-069 and TR-369 report formats

	// TR-069 and TR-369 ParameterPerRow report format columns
	ParameterPerRow_ReportTimestamp = "ReportTimestamp"
	ParameterPerRow_ParameterName   = "ParameterName"
	ParameterPerRow_ParameterValue  = "ParameterValue"
	ParameterPerRow_ParameterType   = "ParameterType"
	// TR-069 and TR-369 ParameterPerRow report format columns
)

type CollectorHandler struct {
	collectorService collectorservices.CollectorService
}

func NewCollectorHandler(collectorService collectorservices.CollectorService) *CollectorHandler {
	return &CollectorHandler{collectorService}
}

func (h *CollectorHandler) Collect(writer http.ResponseWriter, request *http.Request) {
	reportFormat := request.Header.Get("BBF-Report-Format")

	oui := request.URL.Query().Get("oui")
	productClass := request.URL.Query().Get("pc")
	serialNumber := request.URL.Query().Get("sn")

	if reportFormat == ReportFormat_ParameterPerRow {
		bulkData := &collectorservices.CSVBulkDataModel{
			ParameterPerRow: []*collectorservices.ParameterPerRowModel{},
		}

		records, err := csv.NewReader(request.Body).ReadAll()

		if err != nil {
			http.Error(writer, "Bad Request: Invalid CSV format", http.StatusBadRequest)

			return
		}

		fields := map[string]int{}

		for recordIndex, record := range records {
			if recordIndex == 0 {
				for fieldIndex, field := range record {
					fields[field] = fieldIndex
				}
			} else {
				reportTimestamp, err := strconv.ParseInt(record[fields[ParameterPerRow_ReportTimestamp]], 10, 64)

				if err != nil {
					http.Error(writer, "Bad Request: Invalid timestamp format", http.StatusBadRequest)

					return
				}

				parameterType := record[fields[ParameterPerRow_ParameterType]]

				if !collectorservices.IsValidParameterType(parameterType) {
					http.Error(writer, "Bad Request: Invalid parameter type", http.StatusBadRequest)

					return
				}

				parameterPerRow := &collectorservices.ParameterPerRowModel{
					ReportTimestamp: time.Unix(reportTimestamp, 0),
					ParameterName:   record[fields[ParameterPerRow_ParameterName]],
					ParameterValue:  record[fields[ParameterPerRow_ParameterValue]],
					ParameterType:   parameterType,
				}

				bulkData.ParameterPerRow = append(bulkData.ParameterPerRow, parameterPerRow)
			}
		}

		if err := h.collectorService.CollectCSV(request.Context(), oui, productClass, serialNumber, bulkData); err != nil {
			if errors.Is(err, collectorservices.ErrBackpressure) {
				http.Error(writer, "Too Many Requests", http.StatusTooManyRequests)
			} else {
				http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}
	}

	if reportFormat == ReportFormat_ParameterPerColumn {
		http.Error(writer, "Bad Request: Unsupported report format ParameterPerColumn. The supported report formats are ParameterPerRow and NameValuePair.", http.StatusBadRequest)

		return
	}

	if reportFormat == ReportFormat_NameValuePair {
		bulkData := &collectorservices.JSONBulkDataModel{
			NameValuePair: &collectorservices.NameValuePairModel{},
		}

		if err := json.NewDecoder(request.Body).Decode(&bulkData.NameValuePair); err != nil {
			http.Error(writer, "Bad Request: Invalid JSON format", http.StatusBadRequest)

			return
		}

		if err := h.collectorService.CollectJSON(request.Context(), oui, productClass, serialNumber, bulkData); err != nil {
			if errors.Is(err, collectorservices.ErrBackpressure) {
				http.Error(writer, "Too Many Requests", http.StatusTooManyRequests)
			} else {
				http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}
	}

	if reportFormat == ReportFormat_ObjectHierarchy {
		http.Error(writer, "Bad Request: Unsupported report format ObjectHierarchy. The supported report formats are ParameterPerRow and NameValuePair.", http.StatusBadRequest)

		return
	}
}
