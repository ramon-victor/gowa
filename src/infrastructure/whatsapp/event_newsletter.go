package whatsapp

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types/events"
)

// createNewsletterJoinPayload creates a webhook payload for newsletter join events
func createNewsletterJoinPayload(evt *events.NewsletterJoin) map[string]any {
	body := make(map[string]any)

	// Create payload structure matching the expected format
	payload := make(map[string]any)

	// Add newsletter metadata
	payload["newsletter_id"] = evt.ID.String()
	if evt.ThreadMeta.Name.Text != "" {
		payload["name"] = evt.ThreadMeta.Name.Text
	}
	if evt.ThreadMeta.Description.Text != "" {
		payload["description"] = evt.ThreadMeta.Description.Text
	}
	payload["subscriber_count"] = evt.ThreadMeta.SubscriberCount
	payload["verification_state"] = evt.ThreadMeta.VerificationState

	// Wrap in payload structure
	body["payload"] = payload

	// Add metadata for webhook processing
	body["event"] = "newsletter.join"
	body["timestamp"] = time.Now().Format(time.RFC3339)

	return body
}

// createNewsletterLeavePayload creates a webhook payload for newsletter leave events
func createNewsletterLeavePayload(evt *events.NewsletterLeave) map[string]any {
	body := make(map[string]any)

	// Create payload structure matching the expected format
	payload := make(map[string]any)

	// Add newsletter leave information
	payload["newsletter_id"] = evt.ID.String()
	payload["role"] = string(evt.Role)

	// Wrap in payload structure
	body["payload"] = payload

	// Add metadata for webhook processing
	body["event"] = "newsletter.leave"
	body["timestamp"] = time.Now().Format(time.RFC3339)

	return body
}

// createNewsletterMuteChangePayload creates a webhook payload for newsletter mute change events
func createNewsletterMuteChangePayload(evt *events.NewsletterMuteChange) map[string]any {
	body := make(map[string]any)

	// Create payload structure matching the expected format
	payload := make(map[string]any)

	// Add newsletter mute change information
	payload["newsletter_id"] = evt.ID.String()
	payload["mute_state"] = string(evt.Mute)

	// Wrap in payload structure
	body["payload"] = payload

	// Add metadata for webhook processing
	body["event"] = "newsletter.mute.change"
	body["timestamp"] = time.Now().Format(time.RFC3339)

	return body
}

// createNewsletterLiveUpdatePayload creates a webhook payload for newsletter live update events
func createNewsletterLiveUpdatePayload(evt *events.NewsletterLiveUpdate) map[string]any {
	body := make(map[string]any)

	// Create payload structure matching the expected format
	payload := make(map[string]any)

	// Add newsletter live update information
	payload["newsletter_id"] = evt.JID.String()
	payload["message_count"] = len(evt.Messages)

	// Add message summaries if available
	if len(evt.Messages) > 0 {
		var messageSummaries []map[string]any
		for _, msg := range evt.Messages {
			summary := map[string]any{
				"message_id":       msg.MessageID,
				"server_id":        msg.MessageServerID,
				"type":             msg.Type,
				"timestamp":        msg.Timestamp.Format(time.RFC3339),
				"views_count":      msg.ViewsCount,
				"reaction_count":   len(msg.ReactionCounts),
			}
			messageSummaries = append(messageSummaries, summary)
		}
		payload["messages"] = messageSummaries
	}

	// Wrap in payload structure
	body["payload"] = payload

	// Add metadata for webhook processing
	body["event"] = "newsletter.live.update"
	body["timestamp"] = time.Now().Format(time.RFC3339)

	return body
}

// forwardNewsletterJoinToWebhook forwards newsletter join events to the configured webhook URLs
func forwardNewsletterJoinToWebhook(ctx context.Context, evt *events.NewsletterJoin) error {
	webhookService := GetWebhookService()
	if webhookService == nil {
		return nil
	}

	payload := createNewsletterJoinPayload(evt)

	// Use webhook service to submit the event
	if err := webhookService.SubmitWebhook(ctx, "newsletter.join", payload); err != nil {
		return fmt.Errorf("submit newsletter join webhook failed: %w", err)
	}

	logrus.Infof("Newsletter join event forwarded to webhook: %s", evt.ID.String())
	return nil
}

// forwardNewsletterLeaveToWebhook forwards newsletter leave events to the configured webhook URLs
func forwardNewsletterLeaveToWebhook(ctx context.Context, evt *events.NewsletterLeave) error {
	webhookService := GetWebhookService()
	if webhookService == nil {
		return nil
	}

	payload := createNewsletterLeavePayload(evt)

	// Use webhook service to submit the event
	if err := webhookService.SubmitWebhook(ctx, "newsletter.leave", payload); err != nil {
		return fmt.Errorf("submit newsletter leave webhook failed: %w", err)
	}

	logrus.Infof("Newsletter leave event forwarded to webhook: %s (role: %s)", evt.ID.String(), evt.Role)
	return nil
}

// forwardNewsletterMuteChangeToWebhook forwards newsletter mute change events to the configured webhook URLs
func forwardNewsletterMuteChangeToWebhook(ctx context.Context, evt *events.NewsletterMuteChange) error {
	webhookService := GetWebhookService()
	if webhookService == nil {
		return nil
	}

	payload := createNewsletterMuteChangePayload(evt)

	// Use webhook service to submit the event
	if err := webhookService.SubmitWebhook(ctx, "newsletter.mute.change", payload); err != nil {
		return fmt.Errorf("submit newsletter mute change webhook failed: %w", err)
	}

	logrus.Infof("Newsletter mute change event forwarded to webhook: %s (mute: %s)", evt.ID.String(), evt.Mute)
	return nil
}

// forwardNewsletterLiveUpdateToWebhook forwards newsletter live update events to the configured webhook URLs
func forwardNewsletterLiveUpdateToWebhook(ctx context.Context, evt *events.NewsletterLiveUpdate) error {
	webhookService := GetWebhookService()
	if webhookService == nil {
		return nil
	}

	payload := createNewsletterLiveUpdatePayload(evt)

	// Use webhook service to submit the event
	if err := webhookService.SubmitWebhook(ctx, "newsletter.live.update", payload); err != nil {
		return fmt.Errorf("submit newsletter live update webhook failed: %w", err)
	}

	logrus.Infof("Newsletter live update event forwarded to webhook: %s (%d messages)", evt.JID.String(), len(evt.Messages))
	return nil
}

// Chat presence webhook functions

// createChatPresencePayload creates a webhook payload for chat presence events (typing indicators)
func createChatPresencePayload(evt *events.ChatPresence) map[string]any {
	body := make(map[string]any)

	// Create payload structure matching the expected format
	payload := make(map[string]any)

	// Add chat presence information
	payload["chat_id"] = evt.Chat.String()
	payload["sender_id"] = evt.Sender.String()
	payload["state"] = string(evt.State)
	payload["media_type"] = string(evt.Media)

	// Wrap in payload structure
	body["payload"] = payload

	// Add metadata for webhook processing
	body["event"] = "chat.presence"
	body["timestamp"] = time.Now().Format(time.RFC3339)

	return body
}

// forwardChatPresenceToWebhook forwards chat presence events (typing indicators) to the configured webhook URLs
func forwardChatPresenceToWebhook(ctx context.Context, evt *events.ChatPresence) error {
	webhookService := GetWebhookService()
	if webhookService == nil {
		return nil
	}

	payload := createChatPresencePayload(evt)

	// Use webhook service to submit the event
	if err := webhookService.SubmitWebhook(ctx, "chat.presence", payload); err != nil {
		return fmt.Errorf("submit chat presence webhook failed: %w", err)
	}

	logrus.Infof("Chat presence event forwarded to webhook: %s is %s in chat %s",
		evt.Sender.String(), evt.State, evt.Chat.String())
	return nil
}

// Offline sync webhook functions

// createOfflineSyncPreviewPayload creates a webhook payload for offline sync preview events
func createOfflineSyncPreviewPayload(evt *events.OfflineSyncPreview) map[string]any {
	body := make(map[string]any)

	// Create payload structure matching the expected format
	payload := make(map[string]any)

	// Add offline sync preview information
	payload["total_events"] = evt.Total
	payload["app_data_changes"] = evt.AppDataChanges
	payload["messages"] = evt.Messages
	payload["notifications"] = evt.Notifications
	payload["receipts"] = evt.Receipts

	// Wrap in payload structure
	body["payload"] = payload

	// Add metadata for webhook processing
	body["event"] = "offline.sync.preview"
	body["timestamp"] = time.Now().Format(time.RFC3339)

	return body
}

// createOfflineSyncCompletedPayload creates a webhook payload for offline sync completed events
func createOfflineSyncCompletedPayload(evt *events.OfflineSyncCompleted) map[string]any {
	body := make(map[string]any)

	// Create payload structure matching the expected format
	payload := make(map[string]any)

	// Add offline sync completed information
	payload["events_synchronized"] = evt.Count

	// Wrap in payload structure
	body["payload"] = payload

	// Add metadata for webhook processing
	body["event"] = "offline.sync.completed"
	body["timestamp"] = time.Now().Format(time.RFC3339)

	return body
}

// forwardOfflineSyncPreviewToWebhook forwards offline sync preview events to the configured webhook URLs
func forwardOfflineSyncPreviewToWebhook(ctx context.Context, evt *events.OfflineSyncPreview) error {
	webhookService := GetWebhookService()
	if webhookService == nil {
		return nil
	}

	payload := createOfflineSyncPreviewPayload(evt)

	// Use webhook service to submit the event
	if err := webhookService.SubmitWebhook(ctx, "offline.sync.preview", payload); err != nil {
		return fmt.Errorf("submit offline sync preview webhook failed: %w", err)
	}

	logrus.Infof("Offline sync preview event forwarded to webhook: %d total events", evt.Total)
	return nil
}

// forwardOfflineSyncCompletedToWebhook forwards offline sync completed events to the configured webhook URLs
func forwardOfflineSyncCompletedToWebhook(ctx context.Context, evt *events.OfflineSyncCompleted) error {
	webhookService := GetWebhookService()
	if webhookService == nil {
		return nil
	}

	payload := createOfflineSyncCompletedPayload(evt)

	// Use webhook service to submit the event
	if err := webhookService.SubmitWebhook(ctx, "offline.sync.completed", payload); err != nil {
		return fmt.Errorf("submit offline sync completed webhook failed: %w", err)
	}

	logrus.Infof("Offline sync completed event forwarded to webhook: %d events synchronized", evt.Count)
	return nil
}