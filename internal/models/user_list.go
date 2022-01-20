package models

type UserList struct {
	RoomOwner int64   `json:"owner"`
	UserList  []int64 `json:"users"`
}