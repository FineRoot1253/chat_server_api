package fcm

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"google.golang.org/api/option"
)

type Service interface {
	SendMsg(token string, data map[string]string) error
	SendMsgAsMultiCast(roomId string, userStateList []models.UserState, userList models.UserList) error
}

type service struct {
	client *messaging.Client
}

func NewService() (Service, error) {
	opt := option.WithCredentialsFile(utils.FIREBASE_ACCOUNT_FILE_PATH)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Print("error initializing app: ", err)
		return nil, &utils.CommonError{Func: "firebase.NewApp", Data: "", Err: err}
	}

	// Access auth service from the default app
	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Print("error getting Auth client: ", err)
		return nil, &utils.CommonError{Func: "app.Messaging", Data: "", Err: err}
	}
	return &service{client: client}, nil
}

func (service *service) SendMsg(token string, data map[string]string) error {

	log.Print("Token of User : ", token)
	log.Print("Token of User : ", data)

	message := &messaging.Message{
		Data:  data,
		Token: token,
	}

	_, err := service.client.Send(context.Background(), message)
	if err != nil {
		return &utils.CommonError{Func: "SendMsg", Data: "", Err: err}
	}

	return nil

}

func (service *service) SendMsgAsMultiCast(roomId string, userStateList []models.UserState, userList models.UserList) error {

	var tokenList []string

	for _, v := range userStateList {
		log.Print("현재 FCM 보낼 유저 : ", v, " ::: OWNER : ", userList.RoomOwner)
		if v.User_Id != userList.RoomOwner && v.User_FCM_TOKEN != "0" && v.User_FCM_TOKEN != "" {
			log.Print("FCM 보낼 유저 : ", v.User_Id, " ::: OWNER : ", userList.RoomOwner)
			tokenList = append(tokenList, v.User_FCM_TOKEN)
		}
	}

	message := &messaging.MulticastMessage{
		Data:   map[string]string{"room_id": roomId, "msgType": "0"},
		Tokens: tokenList,
	}
	if len(tokenList) != 0 {
		_, err := service.client.SendMulticast(context.Background(), message)
		if err != nil {
			log.Print("FCM ", err)
			return &utils.CommonError{Func: "SendMsg", Data: "", Err: err}
		}
	}

	return nil
}
