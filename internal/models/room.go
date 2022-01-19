package models

import (
	"time"
)

type Room struct {
	Room_Id   int64     `gorm:"column:room_id;primaryKey;autoIncrement:true" json:"room_id"`
	Room_Name string    `gorm:"column:room_name" json:"room_name"`
	CreateAt  time.Time `gorm:"column:createat" json:"createat"`
}

type RoomList struct {
	Room_Id    int64     `gorm:"column:room_id;primaryKey;autoIncrement:true" json:"room_id"`
	Room_Name  string    `gorm:"column:room_name" json:"room_name"`
	Room_Count int64     `gorm:"column:room_count" json:"room_count"`
	CreateAt   time.Time `gorm:"column:createat" json:"createat"`
}

func (Room) TableName() string {
	return "chat_server_dev.room"
}
