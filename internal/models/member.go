package models

import "time"

type Member struct {
	Member_Id int64     `gorm:"column:member_id;primary_key;autoIncrement:true" json:"member_id"`
	Room_Id   int64     `gorm:"column:room_id;" json:"room_id"`
	Room      Room      `gorm:"associationForeignKey:Room_Id;reference:Room_Id"`
	User_Id   int64     `gorm:"column:user_id;" json:"user_id"`
	User      User      `gorm:"associationForeignKey:User_Id;reference:User_Id"`
	CreateAt  time.Time `gorm:"column:createat" json:"createat"`
}

func (Member) TableName() string {
	return "chat_server_dev.member"
}
