package validations

import (
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateCreateWebhook(request *webhook.CreateWebhookRequest) error {
	err := validateWebhookCommon(request.URL, request.Secret, request.Events, request.Description)
	if err != nil {
		return pkgError.ValidationError(err.Error())
	}
	return nil
}

func ValidateUpdateWebhook(request *webhook.UpdateWebhookRequest) error {
	err := validateWebhookCommon(request.URL, request.Secret, request.Events, request.Description)
	if err != nil {
		return pkgError.ValidationError(err.Error())
	}
	return nil
}

func validateWebhookCommon(url, secret string, events []string, description string) error {
	validEvents := strings.Join(webhook.ValidEvents, ", ")
	
	// Validate URL
	if err := validation.Validate(url,
		validation.Required,
		validation.By(validateWebhookURL),
		is.URL,
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
	url, ok := value.(string)
	if !ok {
		return fmt.Errorf("must be a string")
	}
	
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("must start with http:// or https://")
	}
	
	if strings.Contains(url, "localhost") || strings.Contains(url, "127.0.0.1") {
		return fmt.Errorf("localhost and 127.0.0.1 are not allowed for webhooks")
	}
	
	return nil
}