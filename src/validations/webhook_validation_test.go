package validations

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid HTTPS URL with UUID path",
			url:     "https://n8n.axolote.xyz/webhook/f7e14021-516e-4e4c-908a-06358b5cfc26",
			wantErr: false,
		},
		{
			name:    "Valid HTTPS URL with leading space",
			url:     " https://n8n.axolote.xyz/webhook/f7e14021-516e-4e4c-908a-06358b5cfc26",
			wantErr: false,
		},
		{
			name:    "Valid HTTPS URL with trailing space",
			url:     "https://n8n.axolote.xyz/webhook/f7e14021-516e-4e4c-908a-06358b5cfc26 ",
			wantErr: false,
		},
		{
			name:    "Valid HTTP URL",
			url:     "http://example.com/webhook",
			wantErr: false,
		},
		{
			name:    "Invalid URL - missing scheme",
			url:     "n8n.axolote.xyz/webhook",
			wantErr: true,
		},
		{
			name:    "Invalid URL - wrong scheme",
			url:     "ftp://example.com/webhook",
			wantErr: true,
		},
		{
			name:    "Invalid URL - missing host",
			url:     "https:///webhook",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test just the is.URL validator first
			err := validation.Validate(tt.url, is.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("is.URL validator error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test the custom validation function
			err = validateWebhookURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWebhookURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test the combined validation (like in the actual code)
			err = validation.Validate(tt.url,
				validation.Required,
				validation.By(validateWebhookURL),
				is.URL,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("Combined validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}