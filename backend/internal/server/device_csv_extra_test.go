package server

import (
	"strings"
	"testing"
)

func TestParseDeviceCSV_EmptyFile(t *testing.T) {
	_, rowErrors, err := parseDeviceCSV(strings.NewReader(""))
	if err != nil {
		t.Fatalf("parseDeviceCSV() error = %v", err)
	}
	if len(rowErrors) != 1 || rowErrors[0].Message != "empty csv" {
		t.Fatalf("rowErrors = %+v, want empty csv", rowErrors)
	}
}

func TestParseDeviceCSV_DefaultDeviceGroup(t *testing.T) {
	csvText := "device_id,product_code,product_model,hardware_version,current_version,device_group,tags\n" +
		"AMS000003,AMS,V9,1.0,v2.3,,{}\n"

	rows, rowErrors, err := parseDeviceCSV(strings.NewReader(csvText))
	if err != nil {
		t.Fatalf("parseDeviceCSV() error = %v", err)
	}
	if len(rowErrors) != 0 {
		t.Fatalf("rowErrors = %+v", rowErrors)
	}
	if rows[0].DeviceGroup != "default" {
		t.Fatalf("device_group = %q, want default", rows[0].DeviceGroup)
	}
}

func TestParseDeviceCSV_MissingRequiredField(t *testing.T) {
	csvText := "device_id,product_code,product_model,hardware_version,current_version,device_group,tags\n" +
		"AMS000002,,V9,1.0,v2.3,org-1001,{}\n"

	_, rowErrors, err := parseDeviceCSV(strings.NewReader(csvText))
	if err != nil {
		t.Fatalf("parseDeviceCSV() error = %v", err)
	}
	if len(rowErrors) != 1 || !strings.Contains(rowErrors[0].Message, "required") {
		t.Fatalf("rowErrors = %+v, want required field error", rowErrors)
	}
}

func TestValidateDeviceCSVRow_FieldLength(t *testing.T) {
	long := strings.Repeat("a", 65)
	row := deviceCSVRow{
		DeviceID:        long,
		ProductCode:     "AMS",
		ProductModel:    "V9",
		HardwareVersion: "1.0",
	}
	if msg := validateDeviceCSVRow(row); msg != "field length exceeds 64 characters" {
		t.Fatalf("msg = %q, want length error", msg)
	}
}
