package models

type RoomResultData struct {
		Room       Room     `json:"room"`
		MemberList []Member `json:"member_list"`
}