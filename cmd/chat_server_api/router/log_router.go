package router

import (
	"github.com/JunGeunHong1129/chat_server_api/internal/chat_log"
	"github.com/gofiber/fiber/v2"
)

func SetLogRouter(app fiber.Router, chatLogHandler chat_log.Handler ) {

	app.Post("/log/chatSomeThing", chatLogHandler.PublishMessageHandler)
	app.Post("/log/restOfMsg", chatLogHandler.GetRestOfMessageHandler)
	app.Get("/log/unReadMsgCount", chatLogHandler.GetUnReadMsgCountHandler)

}
