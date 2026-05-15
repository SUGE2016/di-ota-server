package server

import (
	"strings"
	"testing"
)

func TestParseDeviceCSVValid(t *testing.T) {
	csvText := "device_id,product_code,product_model,hardware_version,current_version,device_group,tags\n" +
		"AMS000001,AMS,V9,1.0,v2.3,org-1001,\"{\"\"batch\"\":\"\"2026-05\"\"}\"\n"

	rows, rowErrors, err := parseDeviceCSV(strings.NewReader(csvText))
	if err != nil {
		t.Fatalf("parseDeviceCSV() error = %v", err)
	}
	if len(rowErrors) != 0 {
		t.Fatalf("rowErrors = %+v, want none", rowErrors)
	}
	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}
	row := rows[0]
	if row.DeviceID != "AMS000001" || row.ProductCode != "AMS" || row.ProductModel != "V9" || row.HardwareVersion != "1.0" || row.CurrentVersion != "v2.3" || row.DeviceGroup != "org-1001" {
		t.Fatalf("row = %+v", row)
	}
	if string(row.Tags) != `{"batch":"2026-05"}` {
		t.Fatalf("tags = %s", row.Tags)
	}
}

func TestParseDeviceCSVRejectsWrongHeader(t *testing.T) {
	rows, rowErrors, err := parseDeviceCSV(strings.NewReader("device_id,product_code\nAMS000001,AMS\n"))
	if err != nil {
		t.Fatalf("parseDeviceCSV() error = %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("len(rows) = %d, want 0", len(rows))
	}
	if len(rowErrors) != 1 || rowErrors[0].Row != 1 {
		t.Fatalf("rowErrors = %+v, want header error", rowErrors)
	}
}

func TestParseDeviceCSVRejectsDuplicateDeviceID(t *testing.T) {
	csvText := "device_id,product_code,product_model,hardware_version,current_version,device_group,tags\n" +
		"AMS000001,AMS,V9,1.0,v2.3,org-1001,{}\n" +
		"AMS000001,AMS,V9,1.0,v2.3,org-1001,{}\n"

	rows, rowErrors, err := parseDeviceCSV(strings.NewReader(csvText))
	if err != nil {
		t.Fatalf("parseDeviceCSV() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}
	if len(rowErrors) != 1 || !strings.Contains(rowErrors[0].Message, "duplicate device_id") {
		t.Fatalf("rowErrors = %+v, want duplicate device_id", rowErrors)
	}
}

func TestParseDeviceCSVRejectsNonObjectTags(t *testing.T) {
	csvText := "device_id,product_code,product_model,hardware_version,current_version,device_group,tags\n" +
		"AMS000001,AMS,V9,1.0,v2.3,org-1001,[]\n"

	_, rowErrors, err := parseDeviceCSV(strings.NewReader(csvText))
	if err != nil {
		t.Fatalf("parseDeviceCSV() error = %v", err)
	}
	if len(rowErrors) != 1 || rowErrors[0].Message != "tags must be a JSON object" {
		t.Fatalf("rowErrors = %+v, want tags object error", rowErrors)
	}
}
