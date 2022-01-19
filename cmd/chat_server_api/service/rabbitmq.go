package service

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	chat_log "github.com/JunGeunHong1129/chat_server_api/internal/log"
	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"github.com/gomodule/redigo/redis"
	"github.com/streadway/amqp"
)

var RabbitMQChan amqp.Channel

func RabbitMQConnect() {
	conn, err := amqp.Dial("amqp://g9bon:reindeer2017!@haproxy_amqp_lb:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	ch, err1 := conn.Channel()
	if err1 != nil {
		log.Fatal(err1)
	}
	defer ch.Close()
	RabbitMQChan = *ch
	log.Print("RabbitMQ Channel ready")
}

func RabbitMQFirstInit() {

	var roomStates []models.RoomState
	resultSet := service.Connector.Raw("select * from chat_server_dev.room_state where room_state_id  in (select max(room_state_id) from chat_server_dev.room_state group by room_id) and room_state = 1;").Find(&roomStates)

	if resultSet.Error != nil {
		log.Panic(resultSet.Error)
	}
	for _, v := range roomStates {
		strId := strconv.Itoa(int(v.Room_Id))
		RabbitMQChan.QueueDeclare(strId, false, false, false, false, nil)
		errChan := make(chan utils.ErrorStateType, 1)
		go func(id string, errChan chan utils.ErrorStateType) {
			select {
			case <-errChan:
				/// 채팅방 폭파시키든 행위에 대한 리액션 메시지 보내야함
				return
			default:
				errChan <- consumeAndCount(id)
			}
		}(strId, errChan)

	}
}

func consumeAndCount(id string) utils.ErrorStateType {
	log.Print("room " + id + ", consume start")
	msgs, err := RabbitMQChan.Consume(id, "", true, false, false, false, nil)

	if err != nil {
		return utils.Unexpected_Error
	}

	for d := range msgs {

		log.Print("I Recevied SomeThing : ", d.Body, " ::: total : ", len(msgs))
		var pubData chat_log.PublishData
		var chatLogList []models.ChatLog
		var chatStateList []models.ChatState

		redisConn := redisPool.Get()

		if err := json.Unmarshal(d.Body, &pubData); err != nil {

			log.Print("에러발생 : ", err)

			return utils.Unmarshaling_Error
		}

		bytes,err := json.Marshal(&pubData)
		if err != nil {

			log.Print("에러발생 : ", err)

			return utils.Unmarshaling_Error
		}

		log.Print("Recevied Data : ", pubData)
		log.Print("Recevied pubData.ChatContent  : ",pubData.ChatContent)

		log.Print("Recevied pubData.ChatState  : ", pubData.ChatState, " :::: ",strconv.Itoa(pubData.RoomId)+"_u")
		if err := RabbitMQChan.Publish("room_exchange", strconv.Itoa(pubData.RoomId)+"_u", false, false, amqp.Publishing{
			ContentType: "text/json",
			ContentEncoding: "utf-8",
			Body:        bytes}); err != nil {
			log.Print(err)
			return utils.Unexpected_Error

		}
		switch pubData.ChatState {
		/// Chat_State == 2
		case int64(utils.Remove_To_All_Msg):
			res := whenMsgStateDelete(&pubData, redisConn, id, d.Body)
			if res != utils.None {
				return res
			}
			break
			/// Chat_State == 3
		case int64(utils.User_Room_Exit_Msg):
			res := whenMsgStateUserRoomExit(models.Member{Member_Id:int64(pubData.MemberId),Room: models.Room{Room_Id: int64(pubData.RoomId)}})
			if res != utils.None {
				return res
			}
			break
			/// Chat_State == 3
		case int64(utils.User_Room_Add_Msg):
			res := whenMsgStateUserRoomAdd(pubData.ChatContent, int64(pubData.RoomId))
			if res != utils.None {
				return res
			}
			break
		default:
			res := whenMsgStateNormal(&pubData,redisConn, id)
			if res != utils.None {
				return res
			}
			break
		}

		length, err := checkChatListLength(id)
		if err != nil {
			log.Print("에러발생 6 : ", err)
			return utils.Redis_Error
		}

		log.Print("I Check ", id, "`s Length : ", length)
		log.Print("3 DONE")

		/// RDB 저장 로직
		if *length%utils.PER_SAVE_AMOUNT == 0 && *length != 0 {
			log.Print("4")

			val1, err1 := redis.Strings(redisConn.Do("LRANGE", id, 0, -1))
			log.Print("result : ", len(val1))
			if err1 != nil {
				log.Print("에러발생 7 : ", err1)
				return utils.Redis_Error
			}
			val2, err2 := redis.ByteSlices(redisConn.Do("LRANGE", id+"_state", 0, -1))
			log.Print("result : ", len(val2))
			if err2 != nil {
				log.Print("에러발생 8 : ", err2)
				return utils.Redis_Error
			}
			for _, v := range val1 {
				var chatTemp models.ChatLog
				var chatStateTemp models.ChatState
				val3, err3 := redis.Bytes(redisConn.Do("GET", "\""+v+"\""))
				if err3 != nil {
					log.Print("에러 발생 !! : ", err3, " ::: ", v)
					return utils.Redis_Error
				}
				if err := json.Unmarshal(val3, &chatTemp); err != nil {
					log.Print("에러 발생 !! : ", err)
					return utils.Unmarshaling_Error
				}
				if err := json.Unmarshal(val3, &chatStateTemp); err != nil {
					log.Print("에러 발생 !! : ", err)
					return utils.Unmarshaling_Error
				}

				chatLogList = append(chatLogList, chatTemp)
				chatStateList = append(chatStateList, chatStateTemp)
			}
			log.Print("저장전 chatLogList : ", chatLogList)
			/// 중복으로 RDB 테이블에 쌓이는 것을 방지하기 귀찮아서 추가함
			if len(val1) > 5 {
				chatStateList = chatStateList[len(val1)-5:]
			}

			if err := db.Connector.Save(&chatLogList).Error; err != nil {
				log.Print("에러발생 9 : ", err)
				return utils.Rdb_Error
			}
			chatStateNewList := make([]models.ChatState, len(val2))

			for idx, v := range val2 {
				if err := json.Unmarshal(v, &chatStateNewList[idx]); err != nil {
					log.Print("에러발생 11 : ", err)
					return utils.Unmarshaling_Error
				}
				chatStateList = append(chatStateList, chatStateNewList[idx])
			}
			log.Print("저장전 chatStateList : ", chatStateList)

			if err := db.Connector.Save(&chatStateList).Error; err != nil {
				log.Print("에러발생 12 : ", err)
				return utils.Rdb_Error
			}
			///테스트용
			// redisConn.Do("DEL", id)
			if _, err := redisConn.Do("DEL", id+"_state"); err != nil {
				return utils.Redis_Error
			}
			log.Print("4 DONE")

		}
		log.Print("5")
	}
	return utils.None
}

func checkChatListLength(id string) (*int64, error) {
	log.Print("2.1.1")
	redisConn := redisPool.Get()
	log.Print("2.1.1 DONE")
	log.Print("2.1.2")
	res, err := redis.Int64(redisConn.Do("LLen", id))
	if err != nil {
		log.Print("2.1.2 err !!!")
		// log.Panicln("SomeThings Wrong While get length : ", err)
		return nil, err
	}
	log.Print("2.1.2", res)

	return &res, nil
}

// func getChatLogList(id string) (*ChatListModel,error) {
// 	redisConn, err := redisPool.DialContext(backGroundCtx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	val1, err1 := redis.ByteSlices(redisConn.Do("LRANGE", id, 0, -1))
// 	log.Print("result : ", len(val1))
// 	if err1 != nil {
// 		panic(err1)
// 	}

// 	chatLogList := make([]models.ChatLog, len(val1))
// 	chatStateList := make([]models.ChatState, len(val1))

// 	for idx, v := range val1 {
// 		if err2 := json.Unmarshal(v, &chatLogList[idx]); err2 != nil {
// 			log.Panicf("!!! Panic occurred : %v", err2)
// 		}
// 		chatLogList[idx] = models.ChatLog{Chat_Id: chatLogList[idx].Chat_Id, Room: models.Room{Room_Id: chatLogList[idx].Room_Id}, User: models.User{User_Id: chatLogList[idx].User_Id}, Chat_Content: chatLogList[idx].Chat_Content, CreateAt: time.Now()}
// 	}
// 	for idx, v := range val1 {
// 		if err2 := json.Unmarshal(v, &chatStateList[idx]); err2 != nil {
// 			log.Panicf("!!! Panic occurred : %v", err2)
// 		}
// 		chatStateList[idx] = models.ChatState{Chat_Id: chatLogList[idx].Chat_Id, Chat_State: chatStateList[idx].Chat_State, CreateAt: time.Now()}
// 	}
// 	getChatNewStateList(id, &chatStateList)

// 	return &ChatListModel{ChatLogList: chatLogList, ChatStateList: chatStateList},nil
// }

func getChatLogModelList(roomId int, memberId int) ([]models.ChatLogModel, error) {

	var member models.MemberState
	var chatLogList []models.ChatLog
	redisConn := redisPool.Get()
	log.Print("getChatLogModelList 시작")

	if err := db.Connector.Where("member_id = ? and member_state = ?", memberId, 1).Last(&member).Error; err != nil {
		return nil, err
	}
	log.Print("getChatLogModelList RDM 셀렉 끝 ::: ", member.Member_Last_Read_Msg_Index)

	val1, err := redis.Strings(redisConn.Do("LRANGE", roomId, member.Member_Last_Read_Msg_Index, -1))
	if err != nil {
		return nil, err
	}
	for _, v := range val1 {
		var chatTemp models.ChatLog
		log.Print(v, " 조회 시작")
		val2, err := redis.Bytes(redisConn.Do("GET", "\""+v+"\""))
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(val2, &chatTemp); err != nil {
			log.Print("에러 발생 !! : ", err)
			return nil, err
		}

		chatLogList = append(chatLogList, chatTemp)
	}

	chatLogModelList := make([]models.ChatLogModel, len(val1))
	for idx, v := range chatLogList {
		chatLogModelList[idx] = models.ChatLogModel{Chat_Id: v.Chat_Id, Room: models.Room{Room_Id: v.Room_Id}, User: models.User{User_Id: v.User_Id}, Chat_Content: v.Chat_Content, Chat_State: 0, CreateAt: v.CreateAt}
	}
	log.Print("결과 : ", chatLogModelList)

	return chatLogModelList, nil
}

// func getChatNewStateList(id string, chatStateList *[]models.ChatState) {
// 	redisConn, err := redisPool.DialContext(backGroundCtx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	val2, err2 := redis.ByteSlices(redisConn.Do("LRANGE", id+"_state", 0, -1))
// 	log.Print("result : ", len(val2))
// 	if err2 != nil {
// 		panic(err2)
// 	}
// 	chatStateNewList := make([]models.ChatState, len(val2))

// 	for idx, v := range val2 {
// 		if err2 := json.Unmarshal(v, &chatStateNewList[idx]); err2 != nil {
// 			log.Panicf("!!! Panic occurred : %v", err2)
// 		}
// 		log.Print((*chatStateList)[idx].Chat_Id)
// 		*chatStateList = append(*chatStateList, chatStateNewList[idx])
// 	}

// }

func whenMsgStateDelete(pubData *PublishData, redisConn redis.Conn, id string, d []byte) ErrorStateType {
	/// 삭제 요청 pub인지 검사한다. 삭제 요청은 state리스트에만 넣는다.

	/// 업데이트시 기존 chatlog의 index 값저장을 위해 들고 온다.
	log.Print("consume 1 ::: ", pubData.Chat_Id)

	var oldPubData PublishData

	val1, err1 := redis.Bytes(redisConn.Do("GET", "\""+pubData.Chat_Id+"\""))
	if err1 != nil {
		log.Print("에러발생 1 : ", err1)
		return utils.Redis_Error
	}

	if err2 := json.Unmarshal(val1, &oldPubData); err2 != nil {
		log.Print("에러발생 2 : ", err2)
		return utils.Unmarshaling_Error
	}

	/// 신규 chatlog 구조에 추가해준다.
	pubData.List_Index = oldPubData.List_Index

	inputData, err := json.Marshal(pubData)
	if err != nil {
		log.Print("에러발생 3 : ", err)
		return utils.Marshal_Error
	}

	/// 채팅 로그 업데이트
	if _, err := redisConn.Do("SET", pubData.Chat_Id, inputData); err != nil {
		log.Print("에러발생 4 : ", err)
		return utils.Redis_Error
	}

	/// 들어온 상태 메시지 전체를 상태 리스트에 추가
	if _, err := redisConn.Do("RPUSH", id+"_state", d); err != nil {
		log.Print("에러발생 5 : ", err)
		return utils.Redis_Error
	}

	log.Print("1 done")

	return None

}

func whenMsgStateNormal(pubData *PublishData, redisConn redis.Conn, id string) ErrorStateType {

	/// 신규 채팅 로그 생성
	log.Print("2")

	length, err := checkChatListLength(id)
	if err != nil {
		log.Print("에러발생 4 : ", err)
		return utils.Redis_Error
	}
	log.Print("2.1")
	pubData.List_Index = int(*length)
	/// TODO : 원래 이건 클라에서 받아야함, 클라구축시 삭제 필요
	pubData.CreateAt = time.Now()

	log.Print("2.2")

	inputData, err := json.Marshal(pubData)
	if err != nil {
		log.Print("에러발생 5 : ", err)
		return utils.Marshal_Error

	}
	log.Print("2.3")
	n, err := redisConn.Do("SETNX", "\""+pubData.Chat_Id+"\"", inputData)
	if err != nil {
		log.Print("에러발생 6 : ", err)
		return utils.Redis_Error
	}

	log.Print("성공 6 : ", n)

	/// 들어온 채팅의 ID를 룸리스트에 추가
	n2, err1 := redisConn.Do("RPUSH", id, pubData.Chat_Id)
	if err1 != nil {
		log.Print("에러발생 7 : ", err)
		return utils.Redis_Error
	}
	log.Print("성공 7 : ", n2)
	log.Print("2 done")

	log.Print("3")
	return utils.None
}

func whenMsgStateUserRoomExit(member models.Member) utils.ErrorStateType {

	/// 1) 맴버 새 상태 생성
	if err := db.Connector.Create(&models.MemberState{Member: member, Member_State: 0, CreateAt: time.Now()}).Error; err != nil {
		log.Print(err)
		return utils.Rdb_Error
	}

	var userList []models.User
	/// 2) 새 유저 리스트 받음. 이건...의미 있남...?
	if err := db.Connector.Raw("select * from (select * from chat_server_dev.\"member\" where room_id = ?) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1);", member.Room_Id).Scan(&userList).Error; err != nil {
		log.Print(err)
		return utils.Rdb_Error
	}

	/// 3) 유저가 아무도 없는 상태인 경우
	if userList == nil {
		/// 방 폭파 했다고 상태 생성
		if err := db.Connector.Create(&models.RoomState{Room: models.Room{Room_Id: member.Room.Room_Id}, Room_State: 0, CreateAt: time.Now()}).Error; err != nil {
			log.Print("### 에러발생 : RoomState 생성중 ### ")
			log.Print(err)
			return utils.Rdb_Error
		}
		/// 메시지큐 파괴 [누가 컨슘을 하던, 아직 메시지가 남아 있던]
		if _, err := RabbitMQChan.QueueDelete(strconv.Itoa(int(member.Room.Room_Id)), false, false, false); err != nil {
			log.Print("### 에러발생 : QueueDelete중 ###")
			log.Print(err.Error())
			return utils.Unexpected_Error
		}
	}

	return utils.None
}

func whenMsgStateUserRoomAdd(userListStr string, roomId int64) utils.ErrorStateType {
	var userList UserList

	/// ex : `{"users":"[1,2,3]"}` 이런 데이터를 받기를 고대하고 있다.
	if err := json.Unmarshal([]byte(userListStr), &userList); err != nil {
		log.Print(err)
		return utils.Unmarshaling_Error
	}

	/// 맴버 리스트 + 맴버 상태 리스트 
	memberList := make([]models.Member, len(userList.UserList))
	memberStateList := make([]models.MemberState, len(userList.UserList))
	for idx, val := range userList.UserList {
		memberList[idx] = models.Member{Room: models.Room{Room_Id: roomId}, User: models.User{User_Id: val}, CreateAt: time.Now()}
	}

	/// 추가될 맴버 생성
	if err := db.Connector.Create(&memberList).Error; err != nil {
		log.Print(err)
		return utils.Rdb_Error
	}
	for idx, _ := range userList.UserList {

		memberStateList[idx] = models.MemberState{Member: memberList[idx], Member_State: 1, CreateAt: time.Now()}
	}
	/// 추가된 맴버들을 토대로 맴버 상태 생성
	if err := db.Connector.Create(&memberStateList).Error; err != nil {
		log.Print(err)
		return utils.Rdb_Error
	}

	return utils.None
}