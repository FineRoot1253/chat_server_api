package models

import "time"

type ChatState struct {
	Chat_State_Id int64     `gorm:"column:chat_state_id;primary_key;autoIncrement:true" json:"chat_state_id"`
	Chat_Id       string     `gorm:"column:chat_id;" json:"chat_id"`
	ChatLog       ChatLog   `gorm:"ForeignKey:Chat_Id;associationForeignKey:Chat_Id;reference:Chat_Id"`
	Chat_State    int64      `gorm:"column:chat_state" json:"chat_state"`
	CreateAt      time.Time `gorm:"column:createat" json:"createat"`
}

func (ChatState) TableName() string {
	return "chat_log_dev.chat_state"
}
