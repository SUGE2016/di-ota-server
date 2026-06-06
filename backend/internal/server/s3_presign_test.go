package server

import (
	"strings"
	"testing"

	"ota-server/backend/internal/config"
)

func TestBuildS3PresignedDownloadURLUsesPublicHost(t *testing.T) {
	cfg := &config.Config{
		S3: config.S3Config{
			Endpoint:        "http://minio:9000",
			PublicBaseURL:   "http://localhost:9000/ota-packages",
			Region:          "us-east-1",
			Bucket:          "ota-packages",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			SignedURLTTLSec: 600,
		},
	}

	u, err := buildS3PresignedDownloadURL(cfg, "pkg-test")
	if err != nil {
		t.Fatalf("buildS3PresignedDownloadURL() error = %v", err)
	}
	if strings.Contains(u, "minio:9000") {
		t.Fatalf("download URL should not use internal host: %s", u)
	}
	if !strings.HasPrefix(u, "http://localhost:9000/ota-packages/ota/pkg-test") {
		t.Fatalf("unexpected download URL: %s", u)
	}
}

func TestS3PresignEndpointPrefersPublicBaseURL(t *testing.T) {
	host, secure, err := s3PresignEndpoint(&config.Config{
		S3: config.S3Config{
			Endpoint:      "http://minio:9000",
			PublicBaseURL: "https://cdn.example.com/ota-packages",
		},
	})
	if err != nil {
		t.Fatalf("s3PresignEndpoint() error = %v", err)
	}
	if host != "cdn.example.com" || !secure {
		t.Fatalf("got host=%q secure=%v", host, secure)
	}
}
