package user

import (
	"fmt"
	"log"

	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"gorm.io/gorm"
)

type Repository interface {
	GetUserList(nickname string) ([]models.User, error)
	CreateUser(user *models.User) error
	CreateUserState(userState *models.UserState) error
	GetUserWithError(emailAddr string, user *models.User) error
	WithTx(tx *gorm.DB) repository
}

type repository struct {
	conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return repository{conn: conn}
}

func (repository repository) GetUserList(nickname string) ([]models.User, error) {

	var userList []models.User
	keyword := fmt.Sprint("%", nickname, "%")

	if err := repository.conn.Raw("select * from (select * from (select * from chat_server_dev.user_state where user_state_id  in (select max(user_state_id) from chat_server_dev.user_state group by user_id) and user_state > 0) as us , chat_server_dev.\"user\" as u where us.user_id=u.user_id) as mainu  where mainu.nickname like ? and mainu.user_fcm_token != '0' AND mainu.user_fcm_token != '';", keyword).Find(&userList).Error; err != nil {
		log.Print(err)
		return nil, &utils.CommonError{Func: "GetUserList", Data: nickname, Err: err}
	}
	return userList, nil
}

func (repository repository) CreateUser(user *models.User) error {
	if err := repository.conn.Save(user).Error; err != nil {
		log.Print(err)
		return &utils.CommonError{Func: "CreateUser", Data: "", Err: err}
	}
	return nil
}
func (repository repository) CreateUserState(userState *models.UserState) error {
	if err := repository.conn.Create(&userState).Error; err != nil {
		log.Print(err)
		return &utils.CommonError{Func: "CreateUserState", Data: "", Err: err}
	}
	return nil
}

func (repository repository) GetUserWithError(emailAddr string, user *models.User) error {
	return repository.conn.Raw("select * from chat_server_dev.user u where email_addr = ?;", emailAddr).First(&user).Error
}

func (repository repository) WithTx(tx *gorm.DB) repository {
	repository.conn = tx
	return repository
}
