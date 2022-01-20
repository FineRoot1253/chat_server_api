package models

import "time"

type PublishData struct {
	RoomId      int       `json:"room_id"`
	MemberId    int       `json:"member_id"`
	UserId      int       `json:"user_id"`
	ChatState   int64     `json:"chat_state"`
	ChatContent string    `json:"chat_content"`
	Chat_Id     string    `json:"chat_id"`
	List_Index  int       `json:"list_index,omitempty"`
	CreateAt    time.Time `json:"createat,omitempty"`
}