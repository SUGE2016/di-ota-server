package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ota-server/backend/internal/config"
)

type webhookPayload struct {
	Event         string      `json:"event"`
	Device        string      `json:"device_id,omitempty"`
	TaskID        string      `json:"task_id,omitempty"`
	Status        string      `json:"status,omitempty"`
	SourceVersion string      `json:"source_version,omitempty"`
	TargetVersion string      `json:"target_version,omitempty"`
	Detail        interface{} `json:"detail,omitempty"`
}

func sendOTAWebhook(ctx context.Context, cfg *config.Config, p webhookPayload) error {
	body := gin.H{
		"event_id":    "evt-" + uuid.NewString(),
		"event":       p.Event,
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"data": gin.H{
			"device_id":      p.Device,
			"task_id":        p.TaskID,
			"status":         p.Status,
			"source_version": p.SourceVersion,
			"target_version": p.TargetVersion,
			"detail":         p.Detail,
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := strings.TrimSpace(cfg.Integration.WebhookURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if secret := strings.TrimSpace(cfg.Integration.WebhookSecret); secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write(raw)
		req.Header.Set("X-OTA-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
		} else {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			lastErr = fmt.Errorf("webhook status %d", resp.StatusCode)
		}
		time.Sleep(time.Duration(1<<attempt) * time.Second)
	}
	return lastErr
}
