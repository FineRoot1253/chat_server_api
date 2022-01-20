package room

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/JunGeunHong1129/chat_server_api/db"
	"github.com/JunGeunHong1129/chat_server_api/models"
	"github.com/gofiber/fiber/v2"
)

func GetRoomList(c *fiber.Ctx) error {
	key := c.Query("user_id")
	var roomList []models.RoomList
	keyInt, err := strconv.Atoi(key)
	if err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	if err := db.Connector.Raw("select r.room_id , r.room_name , r.createat,(select count(*) from (select * from chat_server_dev.\"member\" where room_id = r.room_id ) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1)) as room_count from chat_server_dev.room as r join (select  room_id  from (select * from chat_server_dev.\"member\" where user_id = ? ) m join (select * from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1) ms  on m.member_id = ms.member_id) rlist on rlist.room_id = r.room_id;", keyInt).Scan(&roomList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	c.Context().Response.Header.Add("Content-Type", "application/json")

	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: roomList})

}

type UserList struct {
	RoomOwner int64   `json:"owner"`
	UserList  []int64 `json:"users"`
}

func CreateRoom(c *fiber.Ctx) error {

	var room models.Room
	var userList UserList

	if err := json.Unmarshal(c.Body(), &room); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}
	if err := json.Unmarshal(c.Body(), &userList); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}


	memberList := make([]models.Member, len(userList.UserList))
	memberStateList := make([]models.MemberState, len(userList.UserList))

	log.Printf("Unmarshal Result : %d", len(userList.UserList))
	log.Print("Unmarshal Result : ", db.Connector)
	log.Print("Unmarshal Result : ", userList.UserList)

	room.CreateAt = time.Now()

	if err := db.Connector.Save(&room).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}
	if err := db.Connector.Save(&models.RoomState{Room: room, Room_State: 1, CreateAt: room.CreateAt}).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

		if _, err := RabbitMQChan.QueueDeclare(strconv.Itoa(int(room.Room_Id)), false, false, false, false, nil); err != nil {
		log.Print("QueueDeclare phase : " + err.Error())
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "채팅룸 생성중 에러가 발생했습니다."})
	}
	if err := RabbitMQChan.QueueBind(
		strconv.Itoa(int(room.Room_Id)),
		strconv.Itoa(int(room.Room_Id)),
		"room_exchange",
		false, nil,
	); err != nil {
		log.Print("QueueBind phase : " + err.Error())
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "채팅룸 생성중 에러가 발생했습니다."})
	}
	log.Print("QueueDeclare 완료")

	/// consume 시작
	errChan := make(chan ErrorStateType, 1)
	go func(id string, errChan chan ErrorStateType) {
		select {
		case <-errChan:
			/// 채팅방 폭파시키든 행위에 대한 리액션 메시지 보내야함
			switch <-errChan {
			case Unexpected_Error:
				return
			}
			return
		default:
			errChan <- consumeAndCount(id)
		}
	}(strconv.Itoa(int(room.Room_Id)), errChan)

	for idx, val := range userList.UserList {
		memberList[idx] = models.Member{Room: room, User: models.User{User_Id: val}, CreateAt: time.Now()}
	}

	if err := db.Connector.Create(&memberList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	for i, _ := range memberList {

		memberStateList[i] = models.MemberState{Member: memberList[i], Member_State: 1, CreateAt: time.Now()}

	}

	if err := db.Connector.Create(&memberStateList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	log.Print("DB Store 완료")

	c.Context().Response.Header.Add("Content-Type", "application/json")

	///TODO : 배포전 이거 주석해제 해야함
	var userStateList []models.UserState

	if err := db.Connector.Raw("select * from chat_server_dev.user_state where user_state_id  in (select max(user_state_id) from (select * from (select * from chat_server_dev.user_state where user_state_id  in (select max(user_state_id) from chat_server_dev.user_state group by user_id) and user_state > 0)as us , (select m.user_id from chat_server_dev.\"member\" m, (select * from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1) as ms where m.member_id = ms.member_id and m.room_id = ?) as mainm where us.user_id = mainm.user_id ) as mainus group by mainus.user_fcm_token);", room.Room_Id).Scan(&userStateList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}
	for _, v := range userStateList {
		log.Print("현재 FCM 보낼 유저 : ", v, " ::: OWNER : ", userList.RoomOwner)
		if v.User_Id != userList.RoomOwner && v.User_FCM_TOKEN != "0" {
			log.Print("FCM 보낼 유저 : ", v.User_Id, " ::: OWNER : ", userList.RoomOwner)
			SendMsg(v.User_FCM_TOKEN, map[string]string{"room_id": strconv.Itoa(int(room.Room_Id)), "msgType": "0"})
		}
	}

	return c.Status(200).JSON(&models.ResultModel{Code: 1, Msg: "ok", Result: struct {
		Room       models.Room     `json:"room"`
		MemberList []models.Member `json:"member_list"`
	}{room, memberList}})

}

func DeleteMemberInRoom(c *fiber.Ctx) error {
	var member models.Member

	if err := json.Unmarshal(c.Body(), &member); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	/// 1) 맴버 새 상태 생성
	if err := db.Connector.Create(&models.MemberState{Member: member, Member_State: 0, CreateAt: time.Now()}).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	var userList []models.User
	/// 2) 새 유저 리스트 받음. 이건...의미 있남...? 
	if err := db.Connector.Raw("select * from (select * from chat_server_dev.\"member\" where room_id = ?) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1);", member.Room_Id).Scan(&userList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	/// 3) 유저가 아무도 없는 상태인 경우
	///TODO 사용자 이탈시 해당 큐에 publish 필요!!
	if userList == nil {
		/// 방 폭파 했다고 상태 생성
		if err := db.Connector.Create(&models.RoomState{Room: models.Room{Room_Id: member.Room_Id}, Room_State: 0, CreateAt: time.Now()}); err != nil {
			log.Print(err)
			return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
		}
		/// 메시지큐 파괴 [누가 컨슘을 하던, 아직 메시지가 남아 있던]
		if _, err := RabbitMQChan.QueueDelete(strconv.Itoa(int(member.Room_Id)), false, false, false); err != nil {
			log.Print(err)
			return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "채팅룸 폭파중 에러가 발생했습니다."})
		}
	}

	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok"})

}

func UpdateLastReadMsgIndex(c *fiber.Ctx) error {

	var body models.MemberState
	var member models.Member

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	body.CreateAt = time.Now()
	log.Print("파씽 : ", body.Member_Id)
	if err := db.Connector.Where("member_id=?", body.Member_Id).Find(&member).Error; err != nil {
		// log.Print(err)
		log.Print("결과에러에요1 : ", err)

		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}
	log.Print("member_id : 1")
	currentCount, err := checkChatListLength(strconv.Itoa(int(member.Room_Id)))
	if err != nil {
		log.Print("결과에러에요2 : ")

		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}
	log.Print("member_id : 2")

	body.Member_Last_Read_Msg_Index = *currentCount
	log.Print("member_id : 3")

	if err := db.Connector.Create(&body).Error; err != nil {
		log.Print(err)
		log.Print("결과에러에요3 : ")
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}
	log.Print("member_id : 4")

	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok"})

}

func AddMemberOnRoom(c *fiber.Ctx) error {

	var room models.Room
	var userList UserList

	if err := json.Unmarshal(c.Body(), &userList); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}
	if err := json.Unmarshal(c.Body(), &room); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	/// 맴버 리스트 + 맴버 상태 리스트 
	memberList := make([]models.Member, len(userList.UserList))
	memberStateList := make([]models.MemberState, len(userList.UserList))
	for idx, val := range userList.UserList {
		memberList[idx] = models.Member{Room: models.Room{Room_Id: room.Room_Id}, User: models.User{User_Id: val}, CreateAt: time.Now()}
	}

	/// 추가될 맴버 생성
	if err := db.Connector.Create(&memberList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}
	for idx, _ := range userList.UserList {

		memberStateList[idx] = models.MemberState{Member: memberList[idx], Member_State: 1, CreateAt: time.Now()}
	}
	/// 추가된 맴버들을 토대로 맴버 상태 생성
	if err := db.Connector.Create(&memberStateList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok"})

}

func GetAddableUserList(c *fiber.Ctx) error {
	key := c.Query("room_id")

	var userList []models.User

	if err := db.Connector.Raw("select * from (select * from chat_server_dev.user_state where user_state_id  in (select max(user_state_id) from chat_server_dev.user_state group by user_id) and user_state > 0) as us ,(select * from chat_server_dev.\"user\" as u1 where u1.user_id not in (select m.user_id from (select * from chat_server_dev.\"member\" where room_id = ?) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1))) as u where u.user_id=us.user_id  and us.user_fcm_token != '0' AND us.user_fcm_token != '';", key).Scan(&userList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}
	c.Context().Response.Header.Add("Content-Type", "application/json")
	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: userList})
}

func GetUserListOfRoom(c *fiber.Ctx) error {

	key := c.Query("room_id")

	var userList []models.UserInRoom

	if err := db.Connector.Raw("select * from chat_server_dev.\"user\" as u, (select * from (select * from chat_server_dev.\"member\" where room_id = ?) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1)) as m where u.user_id = m.user_id;", key).Scan(&userList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
	}

	c.Context().Response.Header.Add("Content-Type", "application/json")
	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: userList})

}
