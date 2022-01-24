package chat_log

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/presenter"
	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/rabbitmq"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	GetRestOfMessageHandler(c *fiber.Ctx) error
	GetUnReadMsgCountHandler(c *fiber.Ctx) error
	PublishMessageHandler(c *fiber.Ctx) error
}

type handler struct {
	chatLogService  Service
	rabbitMQService rabbitmq.Service
}

func NewHandler(chatLogService Service, rabbitMQService rabbitmq.Service) Handler {
	return handler{chatLogService: chatLogService, rabbitMQService: rabbitMQService}
}

func (handler handler) GetRestOfMessageHandler(c *fiber.Ctx) error {

	var data struct {
		RoomId   int `json:"room_id"`
		MemberId int `json:"member_id"`
	}

	log.Print("BODY :", string(c.Body()))

	if err2 := json.Unmarshal(c.Body(), &data); err2 != nil {
		errLog := &utils.CommonError{Func: "Unmarshal", Data: "", Err: err2}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
	}

	chatLogModelList, err := handler.rabbitMQService.GetChatLogModelList(data.RoomId, data.MemberId)

	if err != nil {
		log.Print(err)
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	c.Context().Response.Header.Add("Content-Type", "application/json")
	return c.Status(200).JSON(presenter.Success(chatLogModelList, "ok"))

}

func (handler handler) GetUnReadMsgCountHandler(c *fiber.Ctx) error {
	key := c.Query("room_id")
	chatLogModelList, err := handler.chatLogService.GetUnReadMsgCount(key)
	if err != nil {
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	c.Context().Response.Header.Add("Content-Type", "application/json")
	return c.Status(200).JSON(presenter.Success(chatLogModelList, "ok"))

}

func (handler handler) PublishMessageHandler(c *fiber.Ctx) error {
	var pubData models.PublishData

	if err := json.Unmarshal(c.Body(), &pubData); err != nil {
		errLog := &utils.CommonError{Func: "PublishMessage", Data: "", Err: err}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
	}
	if pubData.ChatState != 2 {
		pubData.Chat_Id = strconv.Itoa(pubData.RoomId) + "_" + strconv.Itoa(pubData.MemberId) + "_" + time.Now().String()
	}

	body, err := json.Marshal(pubData)
	if err != nil {
		errLog := &utils.CommonError{Func: "PublishMessage", Data: "", Err: err}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
	}

	log.Print("PUBDATA : ", pubData.Chat_Id)

	if err := handler.rabbitMQService.PublishMessage(pubData.RoomId, body); err != nil {
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}
	c.Context().Response.Header.Add("Content-Type", "application/json")
	return c.Status(200).JSON(presenter.Success(pubData.Chat_Id, "ok"))
}
