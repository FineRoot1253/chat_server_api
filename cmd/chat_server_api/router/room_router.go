package router

import (
	"github.com/JunGeunHong1129/chat_server_api/internal/room"
	"github.com/gofiber/fiber/v2"
)

func SetRoomRouter(app fiber.Router, handler room.Handler, txMiddleWare fiber.Handler) {

	app.Post("/room/create", txMiddleWare, handler.CreateRoomHandler)
	app.Get("/room/findAddableUserList", handler.GetAddableUserListHandler)
	app.Get("/room/findRoomListOfUser", handler.GetRoomListHandler)
	app.Get("/room/findUserListOfRoom", handler.GetUserListOfRoomHandler)
	app.Post("/room/updateLastReadMsgIdx", handler.UpdateLastReadMsgIndexHandler)

}
