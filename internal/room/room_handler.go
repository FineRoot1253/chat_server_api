package room

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/presenter"
	"github.com/JunGeunHong1129/chat_server_api/internal/fcm"
	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/rabbitmq"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler interface {
	GetRoomListHandler(c *fiber.Ctx) error
	CreateRoomHandler(c *fiber.Ctx) error
	UpdateLastReadMsgIndexHandler(c *fiber.Ctx) error
	GetAddableUserListHandler(c *fiber.Ctx) error
	GetUserListOfRoomHandler(c *fiber.Ctx) error
}

type handler struct {
	roomService     Service
	fcmService      fcm.Service
	rabbitmqService rabbitmq.Service
}

func NewHandler(roomService Service, fcmService fcm.Service, rabbitmqService rabbitmq.Service) Handler {

	return handler{roomService: roomService, fcmService: fcmService, rabbitmqService: rabbitmqService}
}

func (h handler) GetRoomListHandler(c *fiber.Ctx) error {

	/// parsing
	key := c.Query("user_id")
	keyInt, err := strconv.Atoi(key)
	if err != nil {
		errLog := &utils.CommonError{Func: "GetRoomListHandler", Data: key, Err: err}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
	}

	/// use service
	userList, err := h.roomService.GetRoomListOfUser(keyInt)
	if err != nil {
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}
	c.Context().Response.Header.Add("Content-Type", "application/json")

	return c.Status(200).JSON(presenter.Success(userList, "ok"))

}

func (h handler) CreateRoomHandler(c *fiber.Ctx) error {
	// 방 생성
	// 1)  방, 방 상태 생성
	// 2)  맴버, 맴버 상태 생성
	// 3)  중복없이 유저 상태 리스트 조회
	// 4)  rabbitmq 로직
	// 5)  3)에서 조회한 리스트 유저들에게 fcm 전송

	// TODO: handler 로직, 반드시 넣어야 함
	// 1,2,3)은 한 트렌젝션에서 실행되도록 수정할 것
	// 4,5)는 한 서비스 내에서 123) 실행후 실행되도록 보장 해야함
	/// tx Start
	tx := c.Locals("TX").(*gorm.DB)

	/// parsing
	var room models.Room
	var userList models.UserList

	if err := json.Unmarshal(c.Body(), &room); err != nil {
		log.Print(err)
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}
	if err := json.Unmarshal(c.Body(), &userList); err != nil {
		log.Print(err)
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	roomId := strconv.Itoa(int(room.Room_Id))

	/// use service
	roomResultData, userStateList, err := h.roomService.WithTx(tx).CreateRoom(&room, &userList)

	if err != nil {
		tx.Rollback()
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	if err := h.rabbitmqService.CreateQueue(roomId); err != nil {
		tx.Rollback()
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	if err := h.fcmService.SendMsgAsMultiCast(roomId, userStateList, userList); err != nil {
		tx.Rollback()
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}
	c.Context().Response.Header.Add("Content-Type", "application/json")

	log.Print("반환 예정 데이터 : ",roomResultData)

	return c.Status(200).JSON(presenter.Success(roomResultData, "ok"))

}

func (h handler) UpdateLastReadMsgIndexHandler(c *fiber.Ctx) error {

	/// parsing
	var body models.MemberState
	var member models.Member

	if err := c.BodyParser(&body); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	body.CreateAt = time.Now()

	/// use service
	err := h.roomService.GetMember(body, &member)

	roomId := strconv.Itoa(int(member.Room_Id))

	currentCount, err := h.rabbitmqService.CheckChatListLength(roomId)
	if err != nil {
		log.Print("결과에러에요2 : ")
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}
	log.Print("member_id : 2")

	body.Member_Last_Read_Msg_Index = *currentCount

	if err := h.roomService.CreateMemberState(&body); err != nil {
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}
	c.Context().Response.Header.Add("Content-Type", "application/json")

	return c.Status(200).JSON(presenter.Success(nil, "ok"))

}

func (h handler) GetAddableUserListHandler(c *fiber.Ctx) error {

	/// parsing
	key := c.Query("room_id")
	var userList []models.User

	/// use service
	if err := h.roomService.GetAddableUserList(key, &userList); err != nil {
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	return c.Status(200).JSON(presenter.Success(nil, "ok"))

}

func (h handler) GetUserListOfRoomHandler(c *fiber.Ctx) error {

	/// parsing
	key := c.Query("room_id")
	var userList []models.UserInRoom

	/// use service
	if err := h.roomService.GetUserListOfRoom(key, &userList); err != nil {
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}
	c.Context().Response.Header.Add("Content-Type", "application/json")

	return c.Status(200).JSON(presenter.Success(nil, "ok"))

}
