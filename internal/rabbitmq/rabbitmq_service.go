package rabbitmq

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"github.com/streadway/amqp"
)

type Service interface {
	CreateQueue(roomId string) error
	CheckChatListLength(id string) (*int64, error)
	ConsumeAndCount(id string) error
	GetChatLogModelList(roomId int, memberId int) ([]models.ChatLogModel, error)
	PublishMessage(roomId int, body []byte) error
}

type service struct {
	channel    *amqp.Channel
	repository Repository
}

func NewService(repository Repository) (Service, error) {
	conn, err := amqp.Dial("amqp://g9bon:reindeer2017!@localhost:5672/")
	if err != nil {
		return nil, err
	}
	ch, err1 := conn.Channel()
	if err1 != nil {
		return nil, err
	}

	log.Print("RabbitMQ Channel ready")
	go func() {
		<-conn.NotifyClose(make(chan *amqp.Error))
	}()
	return &service{
		channel:    ch,
		repository: repository,
	}, nil
}

func (service *service) RabbitMQFirstInit() error {

	roomStates, err := service.repository.GetRoomList()
	if err != nil {
		return err
	}
	for _, v := range roomStates {
		strId := strconv.Itoa(int(v.Room_Id))
		service.channel.QueueDeclare(strId, false, false, false, false, nil)
		errChan := make(chan error, 1)
		go func(id string, errChan chan error) {
			select {
			case <-errChan:
				/// 채팅방 폭파시키든 행위에 대한 리액션 메시지 보내야함
				return
			default:
				errChan <- service.ConsumeAndCount(id)
			}
		}(strId, errChan)

	}
	return nil
}

func (service *service) ConsumeAndCount(id string) error {
	log.Print("room " + id + ", consume start")
	msgs, err := service.channel.Consume(id, "", true, false, false, false, nil)

	if err != nil {
		return consumeLevelError(id, err)
	}

	for d := range msgs {

		log.Print("I Recevied SomeThing : ", d.Body, " ::: total : ", len(msgs))
		var pubData models.PublishData
		var chatLogList []models.ChatLog
		var chatStateList []models.ChatState

		if err := json.Unmarshal(d.Body, &pubData); err != nil {

			log.Print("에러발생 : ", err)

			return consumeLevelError("", err)
		}

		bytes, err := json.Marshal(&pubData)
		if err != nil {

			log.Print("에러발생 : ", err)

			return consumeLevelError("", err)
		}

		log.Print("Recevied Data : ", pubData)
		log.Print("Recevied pubData.ChatContent  : ", pubData.ChatContent)

		log.Print("Recevied pubData.ChatState  : ", pubData.ChatState, " :::: ", strconv.Itoa(pubData.RoomId)+"_u")

		/// 받은 메시지 처리
		switch pubData.ChatState {
		/// Chat_State == 2
		case int64(utils.Remove_To_All_Msg):
			res := service.whenMsgStateDelete(&pubData, id, d.Body)
			if res != nil {
				return res
			}
			break
			/// Chat_State == 3
		case int64(utils.User_Room_Exit_Msg):
			res := service.whenMsgStateUserRoomExit(models.Member{Member_Id: int64(pubData.MemberId), Room: models.Room{Room_Id: int64(pubData.RoomId)}})
			if res != nil {
				return res
			}
			break
			/// Chat_State == 3
		case int64(utils.User_Room_Add_Msg):
			res := service.whenMsgStateUserRoomAdd(pubData.ChatContent, int64(pubData.RoomId))
			if res != nil {
				return res
			}
			break
		default:
			res := service.whenMsgStateNormal(&pubData, id)
			if res != nil {
				return res
			}
			break
		}

		/// RDB 저장여부 체크를 위한 조회
		length, err := service.CheckChatListLength(id)
		if err != nil {
			log.Print("에러발생 6 : ", err)
			return consumeLevelError("", err)
		}

		log.Print("I Check ", id, "`s Length : ", length)
		log.Print("3 DONE")

		/// RDB 저장 로직
		if *length%utils.PER_SAVE_AMOUNT == 0 && *length != 0 {
			log.Print("4")

			val1, err1 := service.repository.GetChatLogList(id, 0)
			if err1 != nil {
				log.Print("에러발생 7 : ", err1)
				return consumeLevelError("", err1)
			}
			log.Print("result : ", len(val1))
			val2, err2 := service.repository.GetChatLogStateList(id)
			log.Print("result : ", len(val2))
			if err2 != nil {
				log.Print("에러발생 8 : ", err2)
				return consumeLevelError("", err2)
			}
			for _, v := range val1 {
				var chatTemp models.ChatLog
				var chatStateTemp models.ChatState
				val3, err3 := service.repository.GetChatLogData(v)
				if err3 != nil {
					log.Print("에러 발생 !! : ", err3, " ::: ", v)
					return consumeLevelError("", err3)
				}
				if err := json.Unmarshal(val3, &chatTemp); err != nil {
					log.Print("에러 발생 !! : ", err)
					return consumeLevelError("", err)
				}
				if err := json.Unmarshal(val3, &chatStateTemp); err != nil {
					log.Print("에러 발생 !! : ", err)
					return consumeLevelError("", err)
				}

				chatLogList = append(chatLogList, chatTemp)
				chatStateList = append(chatStateList, chatStateTemp)
			}
			log.Print("저장전 chatLogList : ", chatLogList)
			/// 중복으로 RDB 테이블에 쌓이는 것을 방지하기 귀찮아서 추가함
			if len(val1) > 5 {
				chatStateList = chatStateList[len(val1)-5:]
			}
			chatStateNewList := make([]models.ChatState, len(val2))

			for idx, v := range val2 {
				if err := json.Unmarshal(v, &chatStateNewList[idx]); err != nil {
					log.Print("에러발생 11 : ", err)
					return consumeLevelError("", err)
				}
				chatStateList = append(chatStateList, chatStateNewList[idx])
			}

			if err := service.repository.CreateChatLog(chatLogList, chatStateList); err != nil {
				return consumeLevelError("", err)
			}

			///테스트용
			// redisConn.Do("DEL", id)
			/// 21_01_21 까먹고있었는데 기억났다.
			/// 이건 RDB테이블에 중복으로 데이터가 들어가는 것을 방지하기 위해 넣었다.
			/// 이래야 전에 넣었던 상태 데이터 부터 이어서 넣기 때문에 한번 넣으면 초기화개념으로 날린다.
			if err := service.repository.DeleteChatLogStateList(id); err != nil {
				return consumeLevelError("", err)
			}
			log.Print("4 DONE")

		}
		log.Print("5")
		if err := service.channel.Publish("room_exchange", strconv.Itoa(pubData.RoomId)+"_u", false, false, amqp.Publishing{
			ContentType:     "text/json",
			ContentEncoding: "utf-8",
			Body:            bytes}); err != nil {
			log.Print(err)
			return consumeLevelError("", err)

		}
	}
	return nil
}

func (service *service) CheckChatListLength(id string) (*int64, error) {
	log.Print("2.1.1")
	log.Print("2.1.1 DONE")
	log.Print("2.1.2")
	res, err := service.repository.GetChatLogStateListLength(id)
	if err != nil {
		log.Print("2.1.2 err !!!")
		// log.Panicln("SomeThings Wrong While get length : ", err)
		return nil, err
	}
	log.Print("2.1.2", res)

	return res, nil
}

func (service *service) GetChatLogModelList(roomId int, memberId int) ([]models.ChatLogModel, error) {

	var member models.MemberState
	var chatLogList []models.ChatLog
	log.Print("getChatLogModelList 시작")

	if err := service.repository.GetMemberState(member, memberId); err != nil {
		return nil, err
	}
	log.Print("getChatLogModelList RDM 셀렉 끝 ::: ", member.Member_Last_Read_Msg_Index)

	val1, err := service.repository.GetChatLogList(string(roomId), int(member.Member_Last_Read_Msg_Index))
	if err != nil {
		return nil, err
	}
	for _, v := range val1 {
		var chatTemp models.ChatLog
		log.Print(v, " 조회 시작")
		val2, err := service.repository.GetChatLogData(v)
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

func (service *service) whenMsgStateDelete(pubData *models.PublishData, id string, d []byte) error {
	/// 삭제 요청 pub인지 검사한다. 삭제 요청은 state리스트에만 넣는다.

	/// 업데이트시 기존 chatlog의 index 값저장을 위해 들고 온다.
	log.Print("consume 1 ::: ", pubData.Chat_Id)

	var oldPubData models.PublishData

	val1, err := service.repository.GetChatLogData(pubData.Chat_Id)
	if err != nil {
		return &utils.CommonError{Func: "whenMsgStateDelete", Data: "", Err: err}
	}

	if err := json.Unmarshal(val1, &oldPubData); err != nil {
		log.Print("에러발생 1 : ", err)
		return &utils.CommonError{Func: "whenMsgStateDelete", Data: "", Err: err}
	}

	if err2 := json.Unmarshal(val1, &oldPubData); err2 != nil {
		log.Print("에러발생 2 : ", err2)
		return &utils.CommonError{Func: "whenMsgStateDelete", Data: "", Err: err2}
	}

	/// 신규 chatlog 구조에 추가해준다.
	pubData.List_Index = oldPubData.List_Index

	inputData, err := json.Marshal(pubData)
	if err != nil {
		log.Print("에러발생 3 : ", err)
		return &utils.CommonError{Func: "whenMsgStateDelete", Data: "", Err: err}
	}

	/// 채팅 로그 업데이트 + 들어온 상태 메시지 전체를 상태 리스트에 추가 (트랜젝션)
	if err := service.repository.OnDeleteChatLogData(pubData.Chat_Id, inputData, d); err != nil {
		log.Print("에러발생 4 : ", err)
		return &utils.CommonError{Func: "whenMsgStateDelete", Data: "", Err: err}
	}

	log.Print("1 done")

	return nil

}

func (service *service) whenMsgStateNormal(pubData *models.PublishData, id string) error {

	/// 신규 채팅 로그 생성
	log.Print("2")

	length, err := service.CheckChatListLength(id)
	if err != nil {
		log.Print("에러발생 4 : ", err)
		return &utils.CommonError{Func: "whenMsgStateNormal", Data: "", Err: err}
	}
	log.Print("2.1")
	pubData.List_Index = int(*length)
	/// TODO : 원래 이건 클라에서 받아야함, 클라구축시 삭제 필요
	pubData.CreateAt = time.Now()

	log.Print("2.2")

	inputData, err := json.Marshal(pubData)
	if err != nil {
		log.Print("에러발생 5 : ", err)
		return &utils.CommonError{Func: "whenMsgStateNormal", Data: "", Err: err}

	}
	log.Print("2.3")
	if err := service.repository.OnCreateChatLogData(id, pubData.Chat_Id, inputData); err != nil {
		log.Print("에러발생 6 : ", err)
		return &utils.CommonError{Func: "whenMsgStateNormal", Data: "", Err: err}
	}
	log.Print("성공 7 : ")
	log.Print("2 done")

	log.Print("3")
	return nil
}

func (service *service) whenMsgStateUserRoomExit(member models.Member) error {
	return service.repository.OnDeleteMemberFromRoom(member, *service.channel)
}

func (service *service) whenMsgStateUserRoomAdd(userListStr string, roomId int64) error {
	var userList models.UserList

	/// ex : `{"users":"[1,2,3]"}` 이런 데이터를 받기를 고대하고 있다.
	if err := json.Unmarshal([]byte(userListStr), &userList); err != nil {
		log.Print(err)
		return &utils.CommonError{Func: "whenMsgStateUserRoomAdd", Data: "", Err: err}
	}

	/// 맴버 리스트 + 맴버 상태 리스트
	memberList := make([]models.Member, len(userList.UserList))
	memberStateList := make([]models.MemberState, len(userList.UserList))
	for idx, val := range userList.UserList {
		memberList[idx] = models.Member{Room: models.Room{Room_Id: roomId}, User: models.User{User_Id: val}, CreateAt: time.Now()}
	}

	/// 추가될 맴버 생성
	if err := service.repository.OnCreateMemberInRoom(memberList, memberStateList); err != nil {
		log.Print(err)
		return &utils.CommonError{Func: "whenMsgStateUserRoomAdd", Data: "", Err: err}
	}

	return nil
}

func (service *service) PublishMessage(roomId int, body []byte) error {
	if err := service.channel.Publish("", strconv.Itoa(roomId), false, false, amqp.Publishing{
		ContentType: "Application/json",
		Body:        body}); err != nil {
		log.Print(err)
		return &utils.CommonError{Func: "PublishMessage", Data: "", Err: err}
	}
	return nil
}

func (service *service) CreateQueue(roomId string) error {

	if _, err := service.channel.QueueDeclare(roomId, false, false, false, false, nil); err != nil {
		log.Print("QueueDeclare phase : " + err.Error())
		return &utils.CommonError{Func: "CreateQueue", Data: roomId, Err: err}
	}
	// if err := service.channel.QueueBind(roomId, roomId, "room_exchange", false, nil); err != nil {
	// 	log.Print("QueueBind phase : " + err.Error())
	// 	return &utils.CommonError{Func: "CreateQueue", Data: roomId, Err: err}
	// }
	log.Print("QueueDeclare 완료")

	return nil
}

func consumeLevelError(data string, err error) error {
	return &utils.CommonError{Func: "consumeAndCount", Data: data, Err: err}
}
