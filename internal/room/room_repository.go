package room

import (
	"log"
	

	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"gorm.io/gorm"
)

type Repository interface {
	// GetRoomList() ([]models.RoomState, error)
	GetRoomListOfUser(key int) ([]models.RoomList, error)
	CreateRoom(room models.Room, userList models.UserList) (*models.RoomResultData, []models.UserState, error)
	CreateMemberState(body models.MemberState) error

	GetAddableUserList(key string, userList []models.User) error
	GetUserListOfRoom(key string, userList []models.UserInRoom) error

	// CreateMember(memberList []models.Member, memberStateList []models.MemberState) (*models.RoomResultData, error)
	GetMember(member models.Member, memberId int64) error
	WithTx(tx *gorm.DB) repository
	// DeleteMemberInRoom()
	// UpdateLastReadMsgIndex()
	// AddMemberOnRoom()
	// GetAddableUserList()
	// GetUserListOfRoom()
}

type repository struct {
	conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{conn: conn}
}

func (repo repository) GetRoomListOfUser(keyInt int) ([]models.RoomList, error) {

	var roomList []models.RoomList

	if err := repo.conn.Raw("select r.room_id , r.room_name , r.createat,(select count(*) from (select * from chat_server_dev.\"member\" where room_id = r.room_id ) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1)) as room_count from chat_server_dev.room as r join (select  room_id  from (select * from chat_server_dev.\"member\" where user_id = ? ) m join (select * from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1) ms  on m.member_id = ms.member_id) rlist on rlist.room_id = r.room_id;", keyInt).Scan(&roomList).Error; err != nil {
		log.Print(err)
		return nil, err
	}

	return roomList, nil

}

/// 방을 생성합니다.
/// [service] CreatRoom_1
func (repo repository) CreateRoom(room models.Room, userList models.UserList) (*models.RoomResultData, []models.UserState, error) {
	/// 현재 스택 프레임에서 panic이 발생 했는지 검사하고 감지시 롤백처리
	defer func() {
		if r := recover(); r != nil {
			repo.conn.Rollback()
		}
	}()

	memberList := make([]models.Member, len(userList.UserList))
	memberStateList := make([]models.MemberState, len(userList.UserList))
	var userStateList []models.UserState

	if err := repo.conn.Save(&room).Error; err != nil {
		log.Print(err)
		repo.conn.Rollback()
		return nil, nil, err
	}
	if err := repo.conn.Save(&models.RoomState{Room: room, Room_State: 1, CreateAt: room.CreateAt}).Error; err != nil {
		log.Print(err)
		repo.conn.Rollback()
		return nil, nil, err
	}

	if err := repo.conn.Create(&memberList).Error; err != nil {
		log.Print(err)
		repo.conn.Rollback()
		return nil, nil, err
	}

	if err := repo.conn.Create(&memberStateList).Error; err != nil {
		log.Print(err)
		repo.conn.Rollback()
		return nil, nil, err
	}

	if err := repo.conn.Raw("select * from chat_server_dev.user_state where user_state_id  in (select max(user_state_id) from (select * from (select * from chat_server_dev.user_state where user_state_id  in (select max(user_state_id) from chat_server_dev.user_state group by user_id) and user_state > 0)as us , (select m.user_id from chat_server_dev.\"member\" m, (select * from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1) as ms where m.member_id = ms.member_id and m.room_id = ?) as mainm where us.user_id = mainm.user_id ) as mainus group by mainus.user_fcm_token);", room.Room_Id).Scan(&userStateList).Error; err != nil {
		log.Print(err)
		repo.conn.Rollback()
		return nil, nil, err
	}

	if err := repo.conn.Commit().Error; err != nil {
		repo.conn.Rollback()
		return nil, nil, err
	}

	return &models.RoomResultData{
		Room:       room,
		MemberList: memberList,
	}, userStateList, nil

}

func (repo repository) GetMember(member models.Member, memberId int64) error {
	if err := repo.conn.Where("member_id=?", memberId).Find(&member).Error; err != nil {
		log.Print("결과에러에요1 : ", err)
		return &utils.CommonError{Func: "GetMember", Data: "", Err: err}
	}
	return nil
}

func (repo repository) CreateMemberState(body models.MemberState) error {
	if err := repo.conn.Create(&body).Error; err != nil {
		log.Print("결과에러에요3 : ", err)
		return &utils.CommonError{Func: "CreateMemberState", Data: "", Err: err}
	}
	return nil
}

func (repo repository) GetAddableUserList(key string, userList []models.User) error{
	if err := repo.conn.Raw("select * from (select * from chat_server_dev.user_state where user_state_id  in (select max(user_state_id) from chat_server_dev.user_state group by user_id) and user_state > 0) as us ,(select * from chat_server_dev.\"user\" as u1 where u1.user_id not in (select m.user_id from (select * from chat_server_dev.\"member\" where room_id = ?) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1))) as u where u.user_id=us.user_id  and us.user_fcm_token != '0' AND us.user_fcm_token != '';", key).Scan(&userList).Error; err != nil {
		log.Print(err)
		return &utils.CommonError{Func: "GetAddableUserList", Data: "", Err: err}
	}
	return nil
}

func (repo repository) GetUserListOfRoom(key string, userList []models.UserInRoom) error{
	if err := repo.conn.Raw("select * from chat_server_dev.\"user\" as u, (select * from (select * from chat_server_dev.\"member\" where room_id = ?) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1)) as m where u.user_id = m.user_id;", key).Scan(&userList).Error; err != nil {
		log.Print(err)
		return &utils.CommonError{Func: "GetUserListOfRoom", Data: "", Err: err}
	}
	return nil
}
// func AddMemberOnRoom(c *fiber.Ctx) error {

// 	var room models.Room
// 	var userList UserList

// 	if err := json.Unmarshal(c.Body(), &userList); err != nil {
// 		log.Print(err)
// 		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
// 	}
// 	if err := json.Unmarshal(c.Body(), &room); err != nil {
// 		log.Print(err)
// 		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
// 	}

// 	/// 맴버 리스트 + 맴버 상태 리스트
// 	memberList := make([]models.Member, len(userList.UserList))
// 	memberStateList := make([]models.MemberState, len(userList.UserList))
// 	for idx, val := range userList.UserList {
// 		memberList[idx] = models.Member{Room: models.Room{Room_Id: room.Room_Id}, User: models.User{User_Id: val}, CreateAt: time.Now()}
// 	}

// 	/// 추가될 맴버 생성
// 	if err := db.Connector.Create(&memberList).Error; err != nil {
// 		log.Print(err)
// 		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
// 	}
// 	for idx, _ := range userList.UserList {

// 		memberStateList[idx] = models.MemberState{Member: memberList[idx], Member_State: 1, CreateAt: time.Now()}
// 	}
// 	/// 추가된 맴버들을 토대로 맴버 상태 생성
// 	if err := db.Connector.Create(&memberStateList).Error; err != nil {
// 		log.Print(err)
// 		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
// 	}

// 	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok"})

// }

// func GetAddableUserList(c *fiber.Ctx) error {
// 	key := c.Query("room_id")

// 	var userList []models.User

// 	if err := db.Connector.Raw("select * from (select * from chat_server_dev.user_state where user_state_id  in (select max(user_state_id) from chat_server_dev.user_state group by user_id) and user_state > 0) as us ,(select * from chat_server_dev.\"user\" as u1 where u1.user_id not in (select m.user_id from (select * from chat_server_dev.\"member\" where room_id = ?) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1))) as u where u.user_id=us.user_id  and us.user_fcm_token != '0' AND us.user_fcm_token != '';", key).Scan(&userList).Error; err != nil {
// 		log.Print(err)
// 		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
// 	}
// 	c.Context().Response.Header.Add("Content-Type", "application/json")
// 	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: userList})
// }

// func GetUserListOfRoom(c *fiber.Ctx) error {

// 	key := c.Query("room_id")

// 	var userList []models.UserInRoom

// 	if err := db.Connector.Raw("select * from chat_server_dev.\"user\" as u, (select * from (select * from chat_server_dev.\"member\" where room_id = ?) as m where m.member_id in (select member_id from chat_server_dev.member_state where member_state_id  in (select max(member_state_id) from chat_server_dev.member_state group by member_id) and member_state = 1)) as m where u.user_id = m.user_id;", key).Scan(&userList).Error; err != nil {
// 		log.Print(err)
// 		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업중 에러가 발생했습니다."})
// 	}

// 	c.Context().Response.Header.Add("Content-Type", "application/json")
// 	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: userList})

// }

func (repo repository) WithTx(tx *gorm.DB) repository {

	if tx == nil {
		return repo
	}

	repo.conn = tx

	return repo
}
