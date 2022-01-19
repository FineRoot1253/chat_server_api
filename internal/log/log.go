package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/JunGeunHong1129/chat_server_api/db"
	"github.com/JunGeunHong1129/chat_server_api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/streadway/amqp"
)

type CountModel struct {
	Member_read_seq int64 `json:"member_read_seq"`
	Count           int   `json:"logged_count"`
}

type ChatListModel struct {
	ChatLogList   []models.ChatLog
	ChatStateList []models.ChatState
}

type RemoveRequestModel struct {
	RoomId    string `json:"room_id"`
	MemberId  string `json:"member_id"`
	TargetSeq int64  `json:"target_seq"`
	ChatState int64  `json:"chat_state"`
}

type PublishData struct {
	RoomId      int       `json:"room_id"`
	MemberId    int       `json:"member_id"`
	UserId      int       `json:"user_id"`
	ChatState   int64     `json:"chat_state"`
	ChatContent string    `json:"chat_content"`
	Chat_Id     string    `json:"chat_id"`
	List_Index  int       `json:"list_index,omitempty"`
	CreateAt    time.Time `json:"createat,omitempty"`
}

func PublishThis(c *fiber.Ctx) error {

	var pubData PublishData

	if err := json.Unmarshal(c.Body(), &pubData); err != nil {
		log.Print("데이터 파싱중 에러가 발생했습니다.1",err)

		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}
	if pubData.ChatState != 2 {
		pubData.Chat_Id = strconv.Itoa(pubData.RoomId) + "_" + strconv.Itoa(pubData.MemberId) + "_" + time.Now().String()
	}

	body, err := json.Marshal(pubData)
	if err != nil {
		log.Print("데이터 파싱중 에러가 발생했습니다.2",err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	log.Print("PUBDATA : ", pubData.Chat_Id)

	if err := RabbitMQChan.Publish("", strconv.Itoa(pubData.RoomId), false, false, amqp.Publishing{
		ContentType: "Application/json",
		Body:        body}); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "채팅 전송중 에러 발생"})
	}

	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: pubData.Chat_Id})
}

func GetRestOfMessage(c *fiber.Ctx) error {

	var data struct {
		RoomId   int `json:"room_id"`
		MemberId int `json:"member_id"`
	}

	log.Print("BODY :",string(c.Body()))

	if err2 := json.Unmarshal(c.Body(), &data); err2 != nil {
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	// data =	getChatLogList(data)
	// TODO : 여기서부터 수정 필요
	chatLogModelList, err := getChatLogModelList(data.RoomId, data.MemberId)
	if err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	c.Context().Response.Header.Add("Content-Type", "application/json")
	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: chatLogModelList})

}

// func getTotalCount(roomId string, member_id string) CountModel {

// 	var rsModel CountModel

// 	if err := db.Connector.Raw("select member_read_seq ,(select count(*) from chat_log_dev.chat_log where room_id=?) as logged_count from chat_server_dev.member_state ms where member_id = ?;", roomId, member_id).Scan(&rsModel).Error; err != nil {
// 		log.Panic("쿼리중 에러 발생 : ", err)
// 	}

// 	return rsModel

// }

func GetUnReadMsgCount(c *fiber.Ctx) error {
	key := c.Query("room_id")
	var chatLogModelList []models.ChatLogModel

	if err := db.Connector.Raw("select * from chat_log_dev.chat_log_view where room_id = ?;", key).Scan(&chatLogModelList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	c.Context().Response.Header.Add("Content-Type", "application/json")
	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: chatLogModelList})

}

func BindUserQueue(c *fiber.Ctx) error {
	userId:= c.Query("user_id")
	queueName := fmt.Sprintf("mqtt-subscription-client-%vqos0",userId)
	keyInt, err := strconv.Atoi(userId)

	var roomList []models.RoomList

	if err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}


	if err := db.Connector.Raw("select r.room_id , r.room_name , r.createat,(select count(*) from (select * from chat_server_dev.\"member\" where room_id = r.room_id ) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1)) as room_count from chat_server_dev.room as r join (select  room_id  from (select * from chat_server_dev.\"member\" where user_id = ? ) m join (select * from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1) ms  on m.member_id = ms.member_id) rlist on rlist.room_id = r.room_id;", keyInt).Scan(&roomList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	for _, v := range roomList {

		if err := RabbitMQChan.QueueBind(
			queueName,
			strconv.Itoa(int(v.Room_Id))+"_u",
			"room_exchange",
			false, nil,
		);err!=nil{
			log.Print("")
				return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "큐 바인딩중 에러가 발생했습니다. (없는 큐일 확률이 높습니다.)"})

		}
	}

	

	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok"})

}

///테스트 용으로만 둘것 실제로는 mqtt로 chat_state:2로 pub 하면 됨
// func RemoveRequest(c *fiber.Ctx) error {

// 	chatId := c.Query("chat_id")
// 	roomId := c.Query("room_id")

// 	data, err := json.Marshal(models.ChatState{ChatLog: models.ChatLog{Chat_Id: chatId}, Chat_State: 1, CreateAt: time.Now()})
// 	if err != nil {
// 		log.Panic("마샬링 도중 에러 발생", err)
// 	}
// 	// RedisConn.Do("RPUSH", roomId+"_state", data)

// 	RabbitMQChan.Publish("", roomId, false, false, amqp.Publishing{
// 		ContentType: "Application/json",
// 		Body:        c.Body(),
// 	})
// 	return c.SendStatus(200)

// }
