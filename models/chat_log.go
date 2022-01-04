package models

import "time"

type ChatLog struct {
	Chat_Id      string    `gorm:"column:chat_id;primary_key;autoIncrement:true" json:"chat_id"`
	Chat_Content string    `gorm:"column:chat_content" redis:"chat_content" json:"chat_content"`
	Room_Id      int64     `gorm:"column:room_id;" json:"room_id"`
	Room         Room      `gorm:"associationForeignKey:Room_Id;reference:Room_Id"`
	User_Id      int64     `gorm:"column:user_id;" json:"user_id"`
	User         User      `gorm:"associationForeignKey:User_Id;reference:User_Id"`
	CreateAt     time.Time `gorm:"column:createat" json:"createat"`
}

func (ChatLog) TableName() string {
	return "chat_log_dev.chat_log"
}

type ChatLogModel struct {
	Chat_Id      string    `gorm:"column:chat_id;primary_key;" json:"chat_id"`
	Chat_Content string    `gorm:"column:chat_content" json:"chat_content"`
	Room_Id      int64     `gorm:"column:room_id;" json:"room_id"`
	Room         Room      `gorm:"associationForeignKey:Room_Id;reference:Room_Id"`
	Chat_State   int64     `gorm:"column:chat_state;" json:"chat_state"`
	User_Id      int64     `gorm:"column:user_id;" json:"user_id"`
	User         User      `gorm:"associationForeignKey:User_Id;reference:User_Id"`
	CreateAt     time.Time `gorm:"column:createat" json:"createat"`
}

func (ChatLogModel) TableName() string {
	return "chat_log_dev.chat_log_view"
}
