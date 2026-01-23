package whatsapp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
)

func submitWebhook(ctx context.Context, payload map[string]any, url string) error {
	// Configure HTTP client with optional TLS skip verification
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.WhatsappWebhookInsecureSkipVerify,
		},
	}
	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	postBody, err := json.Marshal(payload)
	if err != nil {
		return pkgError.WebhookError(fmt.Sprintf("Failed to marshal body: %v", err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return pkgError.WebhookError(fmt.Sprintf("error when create http object %v", err))
	}

	secretKey := []byte(config.WhatsappWebhookSecret)
	signature, err := utils.GetMessageDigestOrSignature(postBody, secretKey)
	if err != nil {
		return pkgError.WebhookError(fmt.Sprintf("error when create signature %v", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", fmt.Sprintf("sha256=%s", signature))

	req.Body = io.NopCloser(bytes.NewBuffer(postBody))
	resp, err := client.Do(req)
	if err != nil {
		return pkgError.WebhookError(fmt.Sprintf("error when submit webhook: %v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return pkgError.WebhookError(fmt.Sprintf("webhook returned status %d", resp.StatusCode))
	}

	logrus.Infof("Successfully submitted webhook")
	return nil
}
