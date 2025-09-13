package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
)

type WebhookService struct {
	repo webhook.IWebhookRepository
}

var (
	webhookServiceInstance *WebhookService
	webhookServiceOnce     sync.Once
)

func InitWebhookService(repo webhook.IWebhookRepository) {
	webhookServiceOnce.Do(func() {
		webhookServiceInstance = &WebhookService{repo: repo}
	})
}

func GetWebhookService() *WebhookService {
	return webhookServiceInstance
}

func (s *WebhookService) SubmitWebhook(ctx context.Context, event string, payload map[string]any) error {
	if s == nil || s.repo == nil {
		return nil
	}

	webhooks, err := s.repo.FindByEvent(event)
	if err != nil {
		logrus.Errorf("Failed to find webhooks for event %s: %v", event, err)
		return err
	}

	if len(webhooks) == 0 {
		return nil
	}

	logrus.Infof("Forwarding %s event to %d configured webhook(s)", event, len(webhooks))

	var errors []error
	enabledCount := 0
	for _, wh := range webhooks {
		if !wh.Enabled {
			continue
		}
		enabledCount++

		trimmedURL := strings.TrimSpace(wh.URL)
		err := s.submitWebhook(ctx, payload, trimmedURL, wh.Secret)
		if err != nil {
			errors = append(errors, err)
			logrus.Warnf("Failed to submit webhook to %s: %v", trimmedURL, err)
		}
	}

	if len(errors) == enabledCount && len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func (s *WebhookService) submitWebhook(ctx context.Context, payload map[string]any, url string, secret string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	postBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(postBody))
	if err != nil {
		return fmt.Errorf("error when create http object %v", err)
	}

	if secret != "" {
		signature, err := utils.GetMessageDigestOrSignature(postBody, []byte(secret))
		if err != nil {
			return fmt.Errorf("error when create signature %v", err)
		}
		req.Header.Set("X-Hub-Signature-256", fmt.Sprintf("sha256=%s", signature))
		logrus.Debugf("Added signature header for webhook to %s", url)
	} else {
		logrus.Debugf("Skipping signature verification for webhook to %s (no secret configured)", url)
	}

	req.Header.Set("Content-Type", "application/json")

	var attempt int
	var maxAttempts = 5
	var sleepDuration = 1 * time.Second
	var lastErr error

	for attempt = 0; attempt < maxAttempts; attempt++ {
		req.Body = io.NopCloser(bytes.NewBuffer(postBody))
		resp, err := client.Do(req)
		if err == nil {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				resp.Body.Close()
				logrus.Infof("Successfully submitted webhook on attempt %d", attempt+1)
				return nil
			}
			err = fmt.Errorf("webhook returned status %d", resp.StatusCode)
			resp.Body.Close()
		}
		lastErr = err
		logrus.Warnf("Attempt %d to submit webhook failed: %v", attempt+1, err)
		if attempt < maxAttempts-1 {
			time.Sleep(sleepDuration)
			sleepDuration *= 2
		}
	}

	return fmt.Errorf("error when submit webhook after %d attempts: %v", attempt, lastErr)
}
