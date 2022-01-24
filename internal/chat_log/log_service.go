package chat_log

import "github.com/JunGeunHong1129/chat_server_api/internal/models"

type Service interface {
	GetUnReadMsgCount(roomId string) ([]models.ChatLogModel,error) 
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return service{repository:repository}
}

func (service service) GetUnReadMsgCount(roomId string) ([]models.ChatLogModel,error) {
	return service.repository.GetUnReadMsgCount(roomId)
}