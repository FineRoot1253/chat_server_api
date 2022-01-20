package handler

import (
	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/presenter"
	"github.com/JunGeunHong1129/chat_server_api/internal/room"
	"github.com/gofiber/fiber/v2"
)

func CreateRoomHandler(roomService room.Service, fcmService  fcm.Service, rabbitMqService rabbitmq.Service)fiber.Handler{
	return func(c *fiber.Ctx) error {



		return c.Status(200).JSON(presenter.Success)
	}
}