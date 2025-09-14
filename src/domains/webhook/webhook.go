package webhook

import (
	"time"
)

type Webhook struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Secret      string    `json:"secret,omitempty"`
	Events      []string  `json:"events"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateWebhookRequest struct {
	URL         string   `json:"url" validate:"required,url"`
	Secret      string   `json:"secret"`
	Events      []string `json:"events" validate:"required,min=1,dive,oneof=qr pair.success pair.error qr.scanned.without.multidevice connected keepalive.timeout keepalive.restored logged.out stream.replaced manual.login.reconnect temporary.ban connect.failure client.outdated cat.refresh.error stream.error disconnected message message.ack fb.message undecryptable.message history.sync media.retry receipt.delivered receipt.read receipt.read.self receipt.played message.delete message.revoke group group.join group.leave group.promote group.demote group.info group.picture user.about user.picture identity.change privacy.settings presence chat.presence blocklist newsletter.join newsletter.leave newsletter.mute.change newsletter.live.update offline.sync.preview offline.sync.completed"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
}

type UpdateWebhookRequest struct {
	URL         string   `json:"url" validate:"required,url"`
	Secret      string   `json:"secret"`
	Events      []string `json:"events" validate:"required,min=1,dive,oneof=qr pair.success pair.error qr.scanned.without.multidevice connected keepalive.timeout keepalive.restored logged.out stream.replaced manual.login.reconnect temporary.ban connect.failure client.outdated cat.refresh.error stream.error disconnected message message.ack fb.message undecryptable.message history.sync media.retry receipt.delivered receipt.read receipt.read.self receipt.played message.delete message.revoke group group.join group.leave group.promote group.demote group.info group.picture user.about user.picture identity.change privacy.settings presence chat.presence blocklist newsletter.join newsletter.leave newsletter.mute.change newsletter.live.update offline.sync.preview offline.sync.completed"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
}

var ValidEvents = []string{
	// Connection Events
	"qr",
	"pair.success",
	"pair.error",
	"qr.scanned.without.multidevice",
	"connected",
	"keepalive.timeout",
	"keepalive.restored",
	"logged.out",
	"stream.replaced",
	"manual.login.reconnect",
	"temporary.ban",
	"connect.failure",
	"client.outdated",
	"cat.refresh.error",
	"stream.error",
	"disconnected",

	// Message Events
	"message",
	"message.ack",
	"fb.message",
	"undecryptable.message",
	"history.sync",
	"media.retry",
	"receipt.delivered",
	"receipt.read",
	"receipt.read.self",
	"receipt.played",
	"message.delete",
	"message.revoke",

	// Group Events
	"group",
	"group.join",
	"group.leave",
	"group.promote",
	"group.demote",
	"group.info",
	"group.picture",

	// User Events
	"user.about",
	"user.picture",
	"identity.change",
	"privacy.settings",
	"presence",
	"chat.presence",
	"blocklist",

	// Other Events
	"newsletter.join",
	"newsletter.leave",
	"newsletter.mute.change",
	"newsletter.live.update",
	"offline.sync.preview",
	"offline.sync.completed",
}