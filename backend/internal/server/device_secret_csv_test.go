package server

import (
	"strings"
	"testing"
)

func TestParseDeviceSecretCSV_Valid(t *testing.T) {
	csvText := "device_id,device_secret\nAMS000001,secret-a\nAMS000002,secret-b\n"
	rows, errs, err := parseDeviceSecretCSV(strings.NewReader(csvText))
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 0 {
		t.Fatalf("errs = %+v", errs)
	}
	if len(rows) != 2 || rows[0].DeviceSecret != "secret-a" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestParseDeviceSecretCSV_InvalidHeader(t *testing.T) {
	_, errs, err := parseDeviceSecretCSV(strings.NewReader("sn,secret\na,b\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 1 || errs[0].Row != 1 {
		t.Fatalf("errs = %+v", errs)
	}
}

func TestParseDeviceSecretCSV_EmptySecret(t *testing.T) {
	_, errs, err := parseDeviceSecretCSV(strings.NewReader("device_id,device_secret\nAMS000001,\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 1 {
		t.Fatalf("errs = %+v", errs)
	}
}
