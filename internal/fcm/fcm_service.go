package fcm

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"google.golang.org/api/option"
)

type Service interface{

}

type service struct{
	client *messaging.Client
}

func NewService() (Service,error) {
	opt := option.WithCredentialsFile(utils.FIREBASE_ACCOUNT_FILE_PATH)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Print("error initializing app: ", err)
		return nil, &utils.CommonError{Func: "firebase.NewApp", Data: "",Err: err}
	}

	// Access auth service from the default app
	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Print("error getting Auth client: ", err)
		return nil, &utils.CommonError{Func: "app.Messaging", Data: "",Err: err}
	}
	return &service{client: client},nil
}

var FirebaseClient *messaging.Client

func InitFirebase() {
	opt := option.WithCredentialsFile(utils.FIREBASE_ACCOUNT_FILE_PATH)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	// Access auth service from the default app
	FirebaseClient, err = app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

}

func SendMsg(token string, data map[string]string) {

	log.Print("Token of User : ",token);
	log.Print("Token of User : ",data);

	message := &messaging.Message{
		Data:  data,
		Token: token,
	}

	_, err := FirebaseClient.Send(context.Background(), message)
	if err != nil {
		log.Print(err)
	}

}
