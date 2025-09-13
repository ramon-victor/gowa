package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type Webhook struct {
	Service webhook.IWebhookUsecase
}

func InitRestWebhook(app fiber.Router, service webhook.IWebhookUsecase) Webhook {
	handler := Webhook{Service: service}
	
	app.Get("/webhook", handler.GetAllWebhooks)
	app.Get("/webhook/:id", handler.GetWebhook)
	app.Get("/webhook/events", handler.GetAvailableEvents)
	app.Post("/webhook", handler.CreateWebhook)
	app.Put("/webhook/:id", handler.UpdateWebhook)
	app.Delete("/webhook/:id", handler.DeleteWebhook)
	
	return handler
}

func (h *Webhook) CreateWebhook(c *fiber.Ctx) error {
	var request webhook.CreateWebhookRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
			Results: nil,
		})
	}
	
	err := h.Service.CreateWebhook(&request)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
			Results: nil,
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  201,
		Code:    "CREATED",
		Message: "Webhook created successfully",
		Results: nil,
	})
}

func (h *Webhook) GetAllWebhooks(c *fiber.Ctx) error {
	webhooks, err := h.Service.GetAllWebhooks()
	utils.PanicIfNeeded(err)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Webhooks retrieved successfully",
		Results: webhooks,
	})
}

func (h *Webhook) GetWebhook(c *fiber.Ctx) error {
	id := c.Params("id")
	
	wh, err := h.Service.GetWebhookByID(id)
	utils.PanicIfNeeded(err)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Webhook retrieved successfully",
		Results: wh,
	})
}

func (h *Webhook) UpdateWebhook(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var request webhook.UpdateWebhookRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
			Results: nil,
		})
	}
	
	err := h.Service.UpdateWebhook(id, &request)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
			Results: nil,
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Webhook updated successfully",
		Results: nil,
	})
}

func (h *Webhook) GetAvailableEvents(c *fiber.Ctx) error {
	logrus.Debugf("GetAvailableEvents handler called")
	events, err := h.Service.GetAvailableEvents()
	if err != nil {
		logrus.Errorf("GetAvailableEvents service error: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve available events",
			Results: nil,
		})
	}

	logrus.Debugf("GetAvailableEvents returning %d events", len(events))
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Available events retrieved successfully",
		Results: events,
	})
}

func (h *Webhook) DeleteWebhook(c *fiber.Ctx) error {
	id := c.Params("id")
	
	err := h.Service.DeleteWebhook(id)
	utils.PanicIfNeeded(err)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Webhook deleted successfully",
		Results: nil,
	})
}