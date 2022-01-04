package models

import (
	"time"
)

type RoomState struct {
	Room_State_Id int64     `gorm:"column:room_state_id;primary_key;autoIncrement:true" json:"room_state_id"`
	Room_Id       int64     `gorm:"column:room_id;" json:"room_id"`
	Room          Room      `gorm:"associationForeignKey:Room_Id;references:Room_Id"`
	Room_State    int8      `gorm:"column:room_state" json:"room_state"`
	CreateAt      time.Time `gorm:"column:create_at" json:"create_at"`
}

func (RoomState) TableName() string {
	return "chat_server_dev.room_state"
}
