package routes

import (
	"github.com/JunGeunHong1129/chat_server_api/lib"
	"github.com/gofiber/fiber/v2"
)

func InitaliseHandlers() *fiber.App {
	app := fiber.New()
	api := app.Group("/chat")

	v1 := api.Group("/v1", func(c *fiber.Ctx) error { // middleware for /api/v1
		c.Set("Version", "v1")
		return c.Next()
	})
	v1.Get("/user/checkEmail", lib.CheckUserEmailDup)
	v1.Post("/user/create", lib.CreateUser)
	v1.Get("/user/login", lib.UserLogin)
	v1.Get("/user/getUserList", lib.GetUserList)

	v1.Post("/room/create", lib.CreateRoom)
	v1.Get("/room/findAddableUserList", lib.GetAddableUserList)
	v1.Post("/room/addMember", lib.AddMemberOnRoom)
	v1.Post("/room/deleteMember", lib.DeleteMemberInRoom)
	v1.Get("/room/findRoomListOfUser", lib.GetRoomList)
	v1.Get("/room/findUserListOfRoom", lib.GetUserListOfRoom)
	v1.Post("/room/updateLastReadMsgIdx", lib.UpdateLastReadMsgIndex)

	v1.Post("/log/chatSomeThing", lib.PublishThis)
	// v1.Get("/log/deleteMsg", lib.RemoveRequest)
	v1.Post("/log/restOfMsg", lib.GetRestOfMessage)
	v1.Get("/log/unReadMsgCount", lib.GetUnReadMsgCount)
	v1.Get("/log/bindUserQueue", lib.BindUserQueue)

	return app
}

func InitaliseServiceHandlers() *fiber.App {
	app := fiber.New()
	api := app.Group("/serv")

	v1 := api.Group("/v1", func(c *fiber.Ctx) error { // middleware for /api/v1
		c.Set("Version", "v1")
		return c.Next()
	})
	v1.Get("/startConsume/checkEmail")
	v1.Post("/user/create", lib.CreateUser)
	v1.Get("/user/login", lib.UserLogin)
	v1.Get("/user/getUserList", lib.GetUserList)

	return app
}
