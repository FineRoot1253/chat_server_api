package room

import (
	"log"
	"time"

	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"gorm.io/gorm"
)

type Service interface {
	GetRoomListOfUser(key int) ([]models.RoomList, error)
	CreateRoom(room *models.Room, userList *models.UserList) (*models.RoomResultData, []models.UserState, error)
	GetMember(body models.MemberState, member *models.Member) error
	CreateMemberState(body *models.MemberState) error
	GetAddableUserList(key string, userList *[]models.User) error
	GetUserListOfRoom(key string, userList *[]models.UserInRoom) error
	WithTx(tx *gorm.DB) service
}

type service struct {
	repository Repository
}

func NewService(repo Repository) Service {
	return service{repository: repo}
}

func (s service) GetRoomListOfUser(key int) ([]models.RoomList, error) {

	return s.repository.GetRoomListOfUser(key)

}


func (s service) CreateRoom(room *models.Room, userList *models.UserList) (*models.RoomResultData, []models.UserState, error) {
	room.CreateAt = time.Now()

	return s.repository.CreateRoom(room, userList)
}

func (s service) GetMember(body models.MemberState, member *models.Member) error {
	if err := s.repository.GetMember(member, body.Member_Id); err != nil {
		// log.Print(err)
		log.Print("결과에러에요1 : ", err)

		return err
	}
	return nil
}

func (s service) CreateMemberState(body *models.MemberState) error {
	if err := s.repository.CreateMemberState(body); err != nil {
		log.Print("결과에러에요3 : ", err)
		return err
	}
	return nil
}

func (s service) GetAddableUserList(key string, userList *[]models.User) error {
	return s.repository.GetAddableUserList(key,userList)
}

func (s service) GetUserListOfRoom(key string, userList *[]models.UserInRoom) error {
	return s.repository.GetUserListOfRoom(key,userList)
}

func (service service) WithTx(tx *gorm.DB) service {
	service.repository = service.repository.WithTx(tx)
	return service
}
