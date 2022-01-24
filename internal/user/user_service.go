package user

import (
	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"gorm.io/gorm"
)

type Service interface {
	GetUserList(nickname string) ([]models.User, error)
	CreateUser(user *models.User) error
	CreateUserState(userState *models.UserState) error
	GetUserWithError(emailAddr string, user *models.User) error
	WithTx(tx *gorm.DB) service
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return service{repository: repository}
}

func (service service) GetUserList(nickname string) ([]models.User, error) {
	return service.repository.GetUserList(nickname)
}

func (service service) CreateUser(user *models.User) error {
	return service.repository.CreateUser(user)
}

func (service service) CreateUserState(userState *models.UserState) error {
	return service.repository.CreateUserState(userState)
}

func (service service) GetUserWithError(emailAddr string, user *models.User) error {
	return service.repository.GetUserWithError(emailAddr, user)
}

func (service service) WithTx(tx *gorm.DB) service {
	service.repository = service.repository.WithTx(tx)
	return service
}
