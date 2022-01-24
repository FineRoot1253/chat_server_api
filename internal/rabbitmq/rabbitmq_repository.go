package rabbitmq

import (
	"log"
	"strconv"
	"time"

	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"github.com/gomodule/redigo/redis"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type Repository interface {
	GetRoomList() ([]models.RoomState, error)

	InsertChatLogDataToList(chatId string, inputData []byte) error
	GetChatLogData(chatId string) ([]byte, error)
	UpdateChatLogData(chatId string, inputData []byte) error

	CreateChatLog(chatLogList []models.ChatLog, chatStateList []models.ChatState) error
	CreateChatLogIfNotExist(chatLogList string, inputData []byte) error
	GetChatLogList(chatId string, startIndex int) ([]string, error)
	GetChatLogStateListLength(chatId string) (*int64, error)

	CreateChatLogStateData(chatId string, inputData string) error
	GetChatLogStateList(chatId string) ([][]byte, error)
	DeleteChatLogStateList(chatId string) error

	GetMemberState(member models.MemberState, memberId int) error

	OnDeleteChatLogData(chatId string, chatLogData []byte, chatStateData []byte) error
	OnCreateChatLogData(roomId string, chatId string, inputData []byte) error
	OnDeleteMemberFromRoom(member models.Member, ch amqp.Channel) error
	OnCreateMemberInRoom(memberList []models.Member, memberStateList []models.MemberState) error 
}

type repository struct {
	redisConn redis.Conn
	rdbConn   *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	newPool := &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:25000")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
	return &repository{rdbConn: conn, redisConn: newPool.Get()}
}

func (repo *repository) GetRoomList() ([]models.RoomState, error) {

	var roomList []models.RoomState

	if err := repo.rdbConn.Raw("select * from chat_server_dev.room_state where room_state_id  in (select max(room_state_id) from chat_server_dev.room_state group by room_id) and room_state = 1;").Find(&roomList).Error; err != nil {
		log.Print(err)
		return nil, err
	}

	return roomList, nil

}

func (repo *repository) GetChatLogData(chatId string) ([]byte, error) {

	val1, err1 := redis.Bytes(repo.redisConn.Do("GET", "\""+chatId+"\""))
	if err1 != nil {
		log.Print("에러발생 1 : ", err1)
		return nil, &utils.CommonError{Func: "GetChatLogData", Data: chatId, Err: err1}
	}

	return val1, nil

}

func (repo *repository) GetChatLogList(chatId string, startIndex int) ([]string, error) {

	val1, err1 := redis.Strings(repo.redisConn.Do("LRANGE", chatId, startIndex, -1))
	if err1 != nil {
		log.Print("에러발생 1 : ", err1)
		return nil, &utils.CommonError{Func: "GetChatLogList", Data: chatId, Err: err1}
	}

	return val1, nil

}

func (repo *repository) GetChatLogStateList(chatId string) ([][]byte, error) {

	val2, err2 := redis.ByteSlices(repo.redisConn.Do("LRANGE", chatId+"_state", 0, -1))
	if err2 != nil {
		log.Print("에러발생 1 : ", err2)
		return nil, &utils.CommonError{Func: "GetChatLogStateList", Data: chatId, Err: err2}
	}

	return val2, nil

}

func (repo *repository) CreateChatLog(chatLogList []models.ChatLog, chatStateList []models.ChatState) error {

	tx := repo.rdbConn.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(&chatLogList).Error; err != nil {
		log.Print("에러발생 9 : ", err)
		tx.Rollback()
		return &utils.CommonError{Func: "CreateChatLog", Data: "", Err: err}
	}

	if err := tx.Save(&chatStateList).Error; err != nil {
		log.Print("에러발생 12 : ", err)
		tx.Rollback()
		return &utils.CommonError{Func: "CreateChatLog", Data: "", Err: err}
	}

	return tx.Commit().Error
}

func (repo *repository) DeleteChatLogStateList(chatId string) error {

	if _, err := repo.redisConn.Do("DEL", chatId+"_state"); err != nil {
		log.Print("에러발생 1 : ", err)
		return &utils.CommonError{Func: "DeleteChatLogStateList", Data: chatId, Err: err}
	}

	return nil

}

func (repo *repository) GetChatLogStateListLength(chatId string) (*int64, error) {

	res, err := redis.Int64(repo.redisConn.Do("LLen", chatId))
	if err != nil {
		log.Print("에러발생 1 : ", err)
		return nil, &utils.CommonError{Func: "GetChatLogStateListLength", Data: chatId, Err: err}
	}

	return &res, nil

}
func (repo *repository) GetMemberState(member models.MemberState, memberId int) error {

	if err := repo.rdbConn.Where("member_id = ? and member_state = ?", memberId, 1).Last(&member).Error; err != nil {
		log.Print("에러발생 1 : ", err)
		return &utils.CommonError{Func: "GetMeberState", Data: string(memberId), Err: err}
	}

	return nil

}

func (repo *repository) UpdateChatLogData(chatId string, inputData []byte) error {
	if _, err := repo.redisConn.Do("SET", chatId, inputData); err != nil {
		log.Print("에러발생 4 : ", err)
		return &utils.CommonError{Func: "UpdateChatLogData", Data: string(inputData), Err: err}
	}
	return nil
}

func (repo *repository) InsertChatLogDataToList(chatId string, inputData []byte) error {
	if _, err := repo.redisConn.Do("RPUSH", chatId+"_state", inputData); err != nil {
		log.Print("에러발생 5 : ", err)
		return &utils.CommonError{Func: "InsertChatLogDataToList", Data: "", Err: err}
	}
	return nil
}

func (repo *repository) CreateChatLogStateData(chatId string, inputData string) error {
	if _, err1 := repo.redisConn.Do("RPUSH", chatId, inputData); err1 != nil {
		log.Print("에러발생 7 : ", err1)
		return &utils.CommonError{Func: "CreateChatLogStateData", Data: "", Err: err1}
	}
	return nil
}

func (repo *repository) CreateChatLogIfNotExist(chatLogList string, inputData []byte) error {
	if _, err := repo.redisConn.Do("SETNX", "\""+chatLogList+"\"", inputData); err != nil {
		log.Print("에러발생 6 : ", err)
		return &utils.CommonError{Func: "CreateChatLogIfNotExist", Data: "", Err: err}
	}
	return nil
}

/// transection 필요한 레디스 처리
/// 1) 삭제요청
/// 2) 노말 메시지 처리

/// 1) 삭제 요청시
func (repo *repository) OnDeleteChatLogData(chatId string, chatLogData []byte, chatStateData []byte) error {

	repo.redisConn.Send("MULTI")
	/// 채팅 로그 업데이트
	repo.redisConn.Send("SET", chatId, chatLogData)
	/// 들어온 상태 메시지 전체를 상태 리스트에 추가
	repo.redisConn.Send("RPUSH", chatId+"_state", chatStateData)
	if _, err := repo.redisConn.Do("EXEC"); err != nil {
		return &utils.CommonError{Func: "OnDeleteChatLogData", Data: "", Err: err}
	}
	return nil

}

/// 2) 노말 메시지 처리시
func (repo *repository) OnCreateChatLogData(roomId string, chatId string, inputData []byte) error {

	repo.redisConn.Send("MULTI")
	/// 채팅 로그 업데이트
	repo.redisConn.Send("SETNX", "\""+chatId+"\"", inputData)
	/// 들어온 상태 메시지 전체를 상태 리스트에 추가
	repo.redisConn.Send("RPUSH", roomId, chatId)
	if _, err := repo.redisConn.Do("EXEC"); err != nil {
		return &utils.CommonError{Func: "OnCreateChatLogData", Data: "", Err: err}
	}
	return nil
}

/// RDB 트렌젝션 처리
/// 1) 방 나가기 처리
/// 2) 유저 추가하기 처리

func (repo *repository) OnDeleteMemberFromRoom(member models.Member, ch amqp.Channel) error {
	tx := repo.rdbConn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	/// 1) 맴버 새 상태 생성
	if err := tx.Create(&models.MemberState{Member: member, Member_State: 0, CreateAt: time.Now()}).Error; err != nil {
		log.Print(err)
		tx.Rollback()
		return &utils.CommonError{Func: "whenMsgStateUserRoomExit", Data: "", Err: err}
	}
	var userList []models.User
	/// 2) 새 유저 리스트 받음.
	if err := tx.Raw("select * from (select * from chat_server_dev.\"member\" where room_id = ?) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1);", member.Room_Id).Scan(&userList).Error; err != nil {
		log.Print(err)
		tx.Rollback()
		return &utils.CommonError{Func: "whenMsgStateUserRoomExit", Data: "", Err: err}
	}
	/// 3) 유저가 아무도 없는 상태인 경우
	if userList == nil {
		/// 방 폭파 했다고 상태 생성
		if err := tx.Create(&models.RoomState{Room: models.Room{Room_Id: member.Room.Room_Id}, Room_State: 0, CreateAt: time.Now()}).Error; err != nil {
			log.Print("### 에러발생 : RoomState 생성중 ### ")
			log.Print(err)
			tx.Rollback()
			return &utils.CommonError{Func: "whenMsgStateUserRoomExit", Data: "", Err: err}
		}
		/// 메시지큐 파괴 [누가 컨슘을 하던, 아직 메시지가 남아 있던]
		if _, err := ch.QueueDelete(strconv.Itoa(int(member.Room.Room_Id)), false, false, false); err != nil {
			log.Print("### 에러발생 : QueueDelete중 ###")
			log.Print(err.Error())
			return &utils.CommonError{Func: "whenMsgStateUserRoomExit", Data: "", Err: err}
		}
	}

	return tx.Commit().Error
}

func (repo *repository) OnCreateMemberInRoom(memberList []models.Member, memberStateList []models.MemberState) error {

	tx := repo.rdbConn.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	/// 추가될 맴버 생성
	if err := tx.Create(memberList).Error; err != nil {
		log.Print(err)
		tx.Rollback()
		return &utils.CommonError{Func: "OnCreateMemberInRoom", Data: "", Err: err}
	}
	for idx, _ := range memberList {

		memberStateList[idx] = models.MemberState{Member: memberList[idx], Member_State: 1, CreateAt: time.Now()}
	}
	/// 추가된 맴버들을 토대로 맴버 상태 생성
	if err := tx.Create(&memberStateList).Error; err != nil {
		log.Print(err)
		tx.Rollback()
		return &utils.CommonError{Func: "OnCreateMemberInRoom", Data: "", Err: err}
	}

	return tx.Commit().Error
}
