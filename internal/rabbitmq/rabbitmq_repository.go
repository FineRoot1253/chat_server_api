package rabbitmq

import (
	"log"

	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
)

type Repository interface{
		GetRoomList() ([]models.RoomState, error)
		GetChatLogData(chatId string) ([]byte, error)
}

type repository struct {
	redisConn redis.Conn
	rdbConn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	newPool:=&redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "redis:25000")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
	return &repository{rdbConn:conn,redisConn: newPool.Get()}
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
		return nil, &utils.CommonError{Func: "GetChatLogData",Data: chatId,Err: err1}
	}

	return val1, nil

}


