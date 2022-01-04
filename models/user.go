package models

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	User_Id      int64        `gorm:"column:user_id;primary_key;autoIncrement:true" json:"user_id"`
	User_Uuid    string       `gorm:"column:user_uuid" json:"user_uuid"`
	Email_Addr   string       `gorm:"column:email_addr" json:"email_addr"`
	Pwd          string       `gorm:"column:pwd" json:"pwd"`
	NickName     string       `gorm:"column:nickname" json:"nickname"`
	CreateAt     time.Time    `gorm:"column:createat" json:"createat"`
	BirthDate    sql.NullTime `gorm:"column:birthdate" json:"birth_date"`
	Phone_Number string       `gorm:"column:phone_number" json:"phone_number"`
}

type UserInRoom struct {
	User_Id      int64        `gorm:"column:user_id;primary_key;autoIncrement:true" json:"user_id"`
	User_Uuid    string       `gorm:"column:user_uuid" json:"user_uuid"`
	Email_Addr   string       `gorm:"column:email_addr" json:"email_addr"`
	Pwd          string       `gorm:"column:pwd" json:"pwd"`
	Member_Id    int64        `gorm:"column:member_id" json:"member_id"`
	NickName     string       `gorm:"column:nickname" json:"nickname"`
	CreateAt     time.Time    `gorm:"column:createat" json:"createat"`
	BirthDate    sql.NullTime `gorm:"column:birthdate" json:"birth_date"`
	Phone_Number string       `gorm:"column:phone_number" json:"phone_number"`
}

func (User) TableName() string {
	return "chat_server_dev.user"
}

func (c User) String() string {
	return fmt.Sprintf("[%v, %v, %v, %v, %v]", c.User_Uuid, c.Email_Addr, c.Pwd, c.NickName, c.Phone_Number)
}
