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
	Events      []string `json:"events" validate:"required,min=1,dive,oneof=qr pair.success pair.error qr.scanned.without.multidevice connected keepalive.timeout keepalive.restored logged.out stream.replaced manual.login.reconnect temporary.ban connect.failure client.outdated cat.refresh.error stream.error disconnected message.received message.ack media.retry receipt message.delete message.revoke message.reaction call.offer call.accept call.reject call.terminate call.pre.accept call.relay.latency call.transport call.offer.notice unknown.call.event app.state app.state.sync.complete archive clear.chat delete.chat delete.for.me mark.chat.as.read pin star mute label.association.chat label.association.message label.edit group group.join group.leave group.promote group.demote group.info group.picture joined.group user.about user.picture push.name push.name.setting business.name picture user.status.mute identity.change privacy.settings presence chat.presence blocklist unarchive.chats.setting newsletter.join newsletter.leave newsletter.mute.change newsletter.live.update offline.sync.preview offline.sync.completed"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
}

type UpdateWebhookRequest struct {
	URL         string   `json:"url" validate:"required,url"`
	Secret      string   `json:"secret"`
	Events      []string `json:"events" validate:"required,min=1,dive,oneof=qr pair.success pair.error qr.scanned.without.multidevice connected keepalive.timeout keepalive.restored logged.out stream.replaced manual.login.reconnect temporary.ban connect.failure client.outdated cat.refresh.error stream.error disconnected message.received message.ack media.retry receipt message.delete message.revoke message.reaction call.offer call.accept call.reject call.terminate call.pre.accept call.relay.latency call.transport call.offer.notice unknown.call.event app.state app.state.sync.complete archive clear.chat delete.chat delete.for.me mark.chat.as.read pin star mute label.association.chat label.association.message label.edit group group.join group.leave group.promote group.demote group.info group.picture joined.group user.about user.picture push.name push.name.setting business.name picture user.status.mute identity.change privacy.settings presence chat.presence blocklist unarchive.chats.setting newsletter.join newsletter.leave newsletter.mute.change newsletter.live.update offline.sync.preview offline.sync.completed"`
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
	"message.received",
	"message.ack",
	"media.retry",
	"receipt",
	"message.delete",
	"message.revoke",
	"message.reaction",

	// Call Events
	"call.offer",
	"call.accept",
	"call.reject",
	"call.terminate",
	"call.pre.accept",
	"call.relay.latency",
	"call.transport",
	"call.offer.notice",
	"unknown.call.event",

	// App State & Sync Events
	"app.state",
	"app.state.sync.complete",

	// Chat Management Events
	"archive",
	"clear.chat",
	"delete.chat",
	"delete.for.me",
	"mark.chat.as.read",
	"pin",
	"star",
	"mute",

	// Label Events
	"label.association.chat",
	"label.association.message",
	"label.edit",

	// Group Events
	"group",
	"group.join",
	"group.leave",
	"group.promote",
	"group.demote",
	"group.info",
	"group.picture",
	"joined.group",

	// User Events
	"user.about",
	"user.picture",
	"push.name",
	"push.name.setting",
	"business.name",
	"picture",
	"user.status.mute",
	"identity.change",
	"privacy.settings",
	"presence",
	"chat.presence",
	"blocklist",
	"unarchive.chats.setting",

	// Newsletter Events
	"newsletter.join",
	"newsletter.leave",
	"newsletter.mute.change",
	"newsletter.live.update",

	// Offline Sync Events
	"offline.sync.preview",
	"offline.sync.completed",
}
