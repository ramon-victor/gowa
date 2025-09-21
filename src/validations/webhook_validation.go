package validations

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ValidateCreateWebhook(request *webhook.CreateWebhookRequest) error {
	if request == nil {
		return pkgError.ValidationError("request cannot be nil")
	}

	err := validateWebhookCommon(request.URL, request.Secret, request.Events, request.Description)
	if err != nil {
		return pkgError.ValidationError(err.Error())
	}
	return nil
}

func ValidateUpdateWebhook(request *webhook.UpdateWebhookRequest) error {
	if request == nil {
		return pkgError.ValidationError("request cannot be nil")
	}

	err := validateWebhookCommon(request.URL, request.Secret, request.Events, request.Description)
	if err != nil {
		return pkgError.ValidationError(err.Error())
	}
	return nil
}

func validateWebhookCommon(url, secret string, events []string, description string) error {
	validEvents := strings.Join(webhook.ValidEvents, ", ")

	if err := validation.Validate(url,
		validation.Required,
		validation.By(validateWebhookURL),
	); err != nil {
		return err
	}

	// Validate Secret
	if err := validation.Validate(secret,
		validation.Length(0, 255),
	); err != nil {
		return err
	}

	// Validate Events
	if err := validation.Validate(events,
		validation.Required,
		validation.Length(1, 0).Error("at least one event must be selected"),
		validation.Each(
			validation.Required,
			validation.By(func(value interface{}) error {
				event, ok := value.(string)
				if !ok {
					return fmt.Errorf("must be a string")
				}
				for _, validEvent := range webhook.ValidEvents {
					if event == validEvent {
						return nil
					}
				}
				return fmt.Errorf("must be one of: %s", validEvents)
			}),
		),
	); err != nil {
		return err
	}

	// Validate Description
	if err := validation.Validate(description,
		validation.Length(0, 500),
	); err != nil {
		return err
	}

	return nil
}

func validateWebhookURL(value interface{}) error {
	raw, ok := value.(string)
	if !ok {
		return fmt.Errorf("must be a string")
	}
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}
	if u.Scheme == "" {
		return fmt.Errorf("URL scheme is required")
	}
	if u.Host == "" {
		return fmt.Errorf("URL host is required")
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("must use http or https scheme, got: %s", scheme)
	}

	if u.User != nil {
		return fmt.Errorf("user info in URL is not allowed for security reasons")
	}

	return nil
}
