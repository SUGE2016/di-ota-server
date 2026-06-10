package server

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

var deviceSecretCSVHeaders = []string{"device_id", "device_secret"}

const deviceSecretCSVTemplate = "device_id,device_secret\nAMS000001,replace-with-32byte-secret\n"

type deviceSecretCSVRow struct {
	DeviceID     string
	DeviceSecret string
}

func parseDeviceSecretCSV(r io.Reader) ([]deviceSecretCSVRow, []deviceCSVRowError, error) {
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
	if !sameCSVHeader(header, deviceSecretCSVHeaders) {
		return nil, []deviceCSVRowError{{Row: 1, Message: "header must be device_id,device_secret"}}, nil
	}

	rows := make([]deviceSecretCSVRow, 0, 64)
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
		if len(record) != len(deviceSecretCSVHeaders) {
			errorsOut = append(errorsOut, deviceCSVRowError{Row: rowNumber, Message: "column count mismatch"})
			continue
		}

		row := deviceSecretCSVRow{
			DeviceID:     strings.TrimSpace(record[0]),
			DeviceSecret: strings.TrimSpace(record[1]),
		}
		if msg := validateDeviceSecretCSVRow(row); msg != "" {
			errorsOut = append(errorsOut, deviceCSVRowError{Row: rowNumber, Message: msg})
			continue
		}
		if previousRow, ok := seen[row.DeviceID]; ok {
			errorsOut = append(errorsOut, deviceCSVRowError{
				Row:     rowNumber,
				Message: fmt.Sprintf("duplicate device_id, first seen at row %d", previousRow),
			})
			continue
		}
		seen[row.DeviceID] = rowNumber
		rows = append(rows, row)
	}

	return rows, errorsOut, nil
}

func validateDeviceSecretCSVRow(row deviceSecretCSVRow) string {
	if row.DeviceID == "" {
		return "device_id is required"
	}
	if row.DeviceSecret == "" {
		return "device_secret is required"
	}
	if len(row.DeviceID) > 64 {
		return "device_id exceeds 64 characters"
	}
	if len(row.DeviceSecret) > 128 {
		return "device_secret must be <= 128 characters"
	}
	return ""
}
