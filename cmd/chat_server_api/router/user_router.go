package router

import (
	"github.com/JunGeunHong1129/chat_server_api/internal/user"
	"github.com/gofiber/fiber/v2"
)

func SetUserRouter(app fiber.Router, handler user.Handler) {
	
	app.Get("/user/checkEmail", handler.CheckUserEmailDupHandler)
	app.Post("/user/create", handler.CreateUserHandler)
	app.Get("/user/login", handler.UserLoginHandler)
	app.Get("/user/getUserList", handler.GetUserListHandler)

}
