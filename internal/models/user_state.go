package models

import "time"

type UserState struct {
	User_State_Id  int64     `gorm:"column:user_state_id;primary_key;autoIncrement:true" json:"user_state_id"`
	User           User      `gorm:"associationForeignKey:User_Id;reference:User_Id"`
	User_Id        int64     `gorm:"column:user_id;" json:"user_id"`
	User_State     int8      `gorm:"column:user_state" json:"user_state"`
	User_FCM_TOKEN string    `gorm:"column:user_fcm_token" json:"user_fcm_token"`
	CreateAt       time.Time `gorm:"column:create_at" json:"create_at"`
}

func (UserState) TableName() string {
	return "chat_server_dev.user_state"
}
