package send

import "mime/multipart"

type StickerRequest struct {
	BaseRequest
	Sticker       *multipart.FileHeader `json:"sticker" form:"sticker"`
	StickerURL    *string               `json:"sticker_url" form:"sticker_url"`
	PackID        string                `json:"pack_id" form:"pack_id"`
	PackName      string                `json:"pack_name" form:"pack_name"`
	PackPublisher string                `json:"pack_publisher" form:"pack_publisher"`
	Emojis        []string              `json:"emojis" form:"emojis"`
}
