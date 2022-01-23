package router

import "github.com/gofiber/fiber/v2"

func SetLogRouter(app fiber.App) {

	app.Post("/room/create", lib.CreateRoom)
	app.Get("/room/findAddableUserList", lib.GetAddableUserList)
	app.Post("/room/addMember", lib.AddMemberOnRoom)
	app.Post("/room/deleteMember", lib.DeleteMemberInRoom)
	app.Get("/room/findRoomListOfUser", lib.GetRoomList)
	app.Get("/room/findUserListOfRoom", lib.GetUserListOfRoom)
	app.Post("/room/updateLastReadMsgIdx", lib.UpdateLastReadMsgIndex)

}
