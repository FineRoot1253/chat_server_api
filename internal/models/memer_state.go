package models

import "time"

type MemberState struct {
	Member_State_Id            int64     `gorm:"column:member_state_id;primary_key;autoIncrement:true" json:"member_state_id"`
	Member_Id                  int64     `gorm:"column:member_id;" json:"member_id"`
	Member                     Member    `gorm:"associationForeignKey:Member_Id;reference:Member_Id"`
	Member_State               int64     `gorm:"column:member_state" json:"member_state"`
	Member_Last_Read_Msg_Index int64     `gorm:"column:member_last_read_msg_index" json:"member_last_read_msg_index"`
	CreateAt                   time.Time `gorm:"column:create_at" json:"create_at"`
}

func (MemberState) TableName() string {
	return "chat_server_dev.member_state"
}
