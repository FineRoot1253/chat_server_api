package room

import (
	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/presenter"
	"github.com/JunGeunHong1129/chat_server_api/internal/rabbitmq"
	"github.com/JunGeunHong1129/chat_server_api/internal/fcm"
	"github.com/gofiber/fiber/v2"
)

type Handler interface{
	CreateRoomHandler(roomService Service, rabbitmqService rabbitmq.Service, fcmService fcm.Service)fiber.Handler
}

type hander struct {

}

func CreateRoomHandler(roomService Service, rabbitmqService rabbitmq.Service, fcmService fcm.Service)fiber.Handler{
	return func(c *fiber.Ctx) error {

		roomService.CreateRoom()
		rabbitmqService
		
		return c.Status(200).JSON(presenter.Success(,"ok"))
	}
}