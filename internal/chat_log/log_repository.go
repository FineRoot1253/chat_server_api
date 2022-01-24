package chat_log

import (
	"log"

	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	GetUnReadMsgCount(roomId string) ([]models.ChatLogModel, error)
}

type repository struct {
	conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return repository{conn: conn}
}

func (repository repository) GetUnReadMsgCount(roomId string) ([]models.ChatLogModel, error) {
	var chatLogModelList []models.ChatLogModel
	if err := repository.conn.Raw("select * from chat_log_dev.chat_log_view where room_id = ?;", roomId).Scan(&chatLogModelList).Error; err != nil {
		log.Print(err)
		return nil, &utils.CommonError{Func: "GetUnReadMsgCount", Data: roomId, Err: err}
	}

	return chatLogModelList, nil
}
