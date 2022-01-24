package main

import (
	"fmt"
	"log"

	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/config"
	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/router"
	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/service"
	"github.com/JunGeunHong1129/chat_server_api/internal/fcm"
	"github.com/JunGeunHong1129/chat_server_api/internal/rabbitmq"
	"github.com/JunGeunHong1129/chat_server_api/internal/room"
	"github.com/gofiber/fiber/v2"
)

func main() {

}

func initDB() string {
	config :=
		service.Db_Config{
			Host:     config.HOST,
			Port:     config.POSTGRES_PORT,
			User:     config.POSTGRES_USER,
			Password: config.POSTGRES_PWD,
			Db:       config.POSTGRES_DB,
		}

	return service.GetConnConfigs(config)

}

func init() {
	/// postgresql Set
	connectionString := initDB()

	log.Print("Starting the HTTP server on port 50000")

	/// fiber setting
	app := fiber.New()
	api := app.Group("/chat")

	v1 := api.Group("/v1", func(c *fiber.Ctx) error { // middleware for /api/v1
		c.Set("Version", "v1")
		return c.Next()
	})
	connnector, err := service.Connect(connectionString)
	if err != nil {
		panic(err)
	}
	rabbitmqRepository := rabbitmq.NewRepository(connnector)
	rabbitmqService, err := rabbitmq.NewService(rabbitmqRepository)
	if err != nil {
		panic(err)
	}

	
	fcmService,err := fcm.NewService()
	if err != nil {
		panic(err)
	}

	roomRepository := room.NewRepository(connnector)
	roomService := room.NewService(roomRepository)
	roomHandler := room.NewHandler(roomService,fcmService,rabbitmqService)
	router.SetRoomRouter(v1,roomHandler)
	
	// app := routes.InitaliseHandlers()
	/// TODO : ExchangeDeclare 선언 필요 위치 FanOut이나 direct

	/// api server start
	log.Fatal(app.Listen(fmt.Sprintf(":%v", config.CHAT_SERVER_PORT)))
}
