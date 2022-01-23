package main

import (
	"fmt"
	"log"

	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/router"
	"github.com/JunGeunHong1129/chat_server_api/db"
	"github.com/JunGeunHong1129/chat_server_api/internal/room"
	"github.com/JunGeunHong1129/chat_server_api/lib"
	"github.com/JunGeunHong1129/chat_server_api/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/streadway/amqp"
)

func main() {

}

func initDB() string {
	config :=
		db.Db_Config{
			Host:     lib.HOST,
			Port:     lib.POSTGRES_PORT,
			User:     lib.POSTGRES_USER,
			Password: lib.POSTGRES_PWD,
			Db:       lib.POSTGRES_DB,
		}

	return db.GetConnConfigs(config)
	

}

func init() {
	/// postgresql Set
	connectionString := initDB()

	/// rabbitmq connection start
	conn, err := amqp.Dial("amqp://g9bon:reindeer2017!@haproxy_amqp_lb:5672/")
	if err != nil {
		log.Fatal(err)
	}
	// defer conn.Close()
	ch, err1 := conn.Channel()
	if err1 != nil {
		log.Fatal(err1)
	}
	// defer ch.Close()
	lib.RabbitMQChan = *ch
	go func() {
		<-conn.NotifyClose(make(chan *amqp.Error))
	}()
	lib.RabbitMQFirstInit()

	log.Print("RabbitMQ Channel ready")

	// /// redis connection start
	// redisConn, err := redis.Dial("tcp", ":"+strconv.Itoa(lib.REDIS_PORT))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer conn.Close()

	// lib.RedisConn = redisConn
	// log.Print("Redis Connection ready")

	lib.InitFirebase() 
	log.Print("Firebase ready")


	log.Print("Starting the HTTP server on port 50000")

	/// fiber setting
	app := fiber.New()
	api := app.Group("/chat")

	v1 := api.Group("/v1", func(c *fiber.Ctx) error { // middleware for /api/v1
		c.Set("Version", "v1")
		return c.Next()
	})
	connnector,err := db.Connect(connectionString)
	if err != nil {
		panic(err)
	}
	roomRepository := room.NewRepository(connnector)
	roomService := room.NewService(roomRepository)
	roomHandler := room.NewHandler()
	router.SetRoomRouter(v1,)
	// app := routes.InitaliseHandlers()
	/// TODO : ExchangeDeclare 선언 필요 위치 FanOut이나 direct

	/// api server start
	log.Fatal(app.Listen(fmt.Sprintf(":%v", lib.CHAT_SERVER_PORT)))
}
