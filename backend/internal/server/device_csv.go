package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var deviceCSVHeaders = []string{
	"device_id",
	"product_code",
	"product_model",
	"hardware_version",
	"current_version",
	"device_group",
	"tags",
}

const deviceCSVTemplate = "device_id,product_code,product_model,hardware_version,current_version,device_group,tags\nAMS000001,AMS,V9,1.0,v2.3,org-1001,\"{\"\"batch\"\":\"\"2026-05\"\"}\"\n"

type deviceCSVRow struct {
	DeviceID        string
	ProductCode     string
	ProductModel    string
	HardwareVersion string
	CurrentVersion  string
	DeviceGroup     string
	Tags            json.RawMessage
}

type deviceCSVRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

func parseDeviceCSV(r io.Reader) ([]deviceCSVRow, []deviceCSVRowError, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err == io.EOF {
		return nil, []deviceCSVRowError{{Row: 1, Message: "empty csv"}}, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read csv header: %w", err)
	}
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}
	if !sameCSVHeader(header, deviceCSVHeaders) {
		return nil, []deviceCSVRowError{{Row: 1, Message: "header must be device_id,product_code,product_model,hardware_version,current_version,device_group,tags"}}, nil
	}

	rows := make([]deviceCSVRow, 0, 64)
	errorsOut := make([]deviceCSVRowError, 0)
	seen := map[string]int{}
	rowNumber := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNumber++
		if err != nil {
			return nil, nil, fmt.Errorf("read csv row %d: %w", rowNumber, err)
		}
		if isBlankCSVRecord(record) {
			continue
		}
		if len(record) != len(deviceCSVHeaders) {
			errorsOut = append(errorsOut, deviceCSVRowError{Row: rowNumber, Message: "column count mismatch"})
			continue
		}

		row := deviceCSVRow{
			DeviceID:        strings.TrimSpace(record[0]),
			ProductCode:     strings.TrimSpace(record[1]),
			ProductModel:    strings.TrimSpace(record[2]),
			HardwareVersion: strings.TrimSpace(record[3]),
			CurrentVersion:  strings.TrimSpace(record[4]),
			DeviceGroup:     strings.TrimSpace(record[5]),
		}
		if row.DeviceGroup == "" {
			row.DeviceGroup = "default"
		}

		if msg := validateDeviceCSVRow(row); msg != "" {
			errorsOut = append(errorsOut, deviceCSVRowError{Row: rowNumber, Message: msg})
			continue
		}
		if previousRow, ok := seen[row.DeviceID]; ok {
			errorsOut = append(errorsOut, deviceCSVRowError{Row: rowNumber, Message: fmt.Sprintf("duplicate device_id, first seen at row %d", previousRow)})
			continue
		}

		tags, msg := normalizeDeviceCSVTags(record[6])
		if msg != "" {
			errorsOut = append(errorsOut, deviceCSVRowError{Row: rowNumber, Message: msg})
			continue
		}
		row.Tags = tags
		seen[row.DeviceID] = rowNumber
		rows = append(rows, row)
	}

	return rows, errorsOut, nil
}

func sameCSVHeader(got []string, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if strings.TrimSpace(got[i]) != want[i] {
			return false
		}
	}
	return true
}

func isBlankCSVRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

func validateDeviceCSVRow(row deviceCSVRow) string {
	if row.DeviceID == "" || row.ProductCode == "" || row.ProductModel == "" || row.HardwareVersion == "" {
		return "device_id/product_code/product_model/hardware_version are required"
	}
	if len(row.DeviceID) > 64 || len(row.ProductCode) > 64 || len(row.ProductModel) > 64 || len(row.HardwareVersion) > 64 || len(row.CurrentVersion) > 64 || len(row.DeviceGroup) > 64 {
		return "field length exceeds 64 characters"
	}
	return ""
}

func normalizeDeviceCSVTags(raw string) (json.RawMessage, string) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return json.RawMessage(`{}`), ""
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(value), &obj); err != nil {
		return nil, "tags must be a JSON object"
	}
	if obj == nil {
		return nil, "tags must be a JSON object"
	}
	normalized, err := json.Marshal(obj)
	if err != nil {
		return nil, "tags must be a JSON object"
	}
	return json.RawMessage(normalized), ""
}
