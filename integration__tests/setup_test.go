package integrestion__tests

import (
	// "fmt"
	"log"

	"github.com/JunGeunHong1129/chat_server_api/db"
	"github.com/JunGeunHong1129/chat_server_api/lib"

	// "github.com/JunGeunHong1129/chat_server_api/routes"
	"github.com/streadway/amqp"
)

/// TODO :
// import (
// 	"fmt"
// 	"log"

// 	"github.com/JunGeunHong1129/chat_server_api/lib"
// )

// func init() {

// 	initDB()
// 	/// rabbitmq connection start
// 	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/chat")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer conn.Close()
// 	ch, err1 := conn.Channel()
// 	if err1 != nil {
// 		log.Fatal(err1)
// 	}
// 	defer ch.Close()
// 	lib.RabbitMQChan = *ch

// 	lib.RabbitMQFirstInit()

// 	log.Print("RabbitMQ Channel ready")

// 	/// redis connection start
// 	redisConn, err := redis.Dial("tcp", ":"+strconv.Itoa(lib.REDIS_PORT))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer conn.Close()

// 	lib.RedisConn = redisConn
// 	log.Print("Redis Connection ready")
// 	log.Println("Starting the HTTP server on port 50000")

// 	app := routes.InitaliseHandlers()
// 	/// TODO : ExchangeDeclare 선언 필요 위치 FanOut이나 direct

// 	log.Fatal(app.Listen(fmt.Sprintf(":%v", lib.CHAT_SERVER_PORT)))
// }
// func TestMain (m *testing.M){
/// Test 시작전 초기화 단계
/// 1) Postgresql
/// 2) Redis
/// 3) RabbitMq
/// 초기화 통합 테스트는 init_integrestion_test.go에서 확인할 것
// initFunc()

/// Test 시작
// os.Exit(m.Run())

// }
func initDB() {
	config :=
		db.Db_Config{
			Host:     lib.HOST,
			Port:     lib.POSTGRES_PORT,
			User:     lib.POSTGRES_USER,
			Password: lib.POSTGRES_PWD,
			Db:       lib.POSTGRES_DB,
		}

	connectionString := db.GetConnConfigs(config)
	err := db.Connect(connectionString)
	if err != nil {
		panic(err.Error())
	}

}

func initFunc() {

	initDB()
	/// rabbitmq connection start
	conn, err := amqp.Dial("amqp://g9bon:reindeer2017!@haproxy_amqp_lb:5672/")
	if err != nil {
		log.Fatal(err)
	}

	// defer conn.Close()

	/// rabbitmq channel start
	ch, err1 := conn.Channel()
	if err1 != nil {
		log.Fatal(err1)
	}

	// defer ch.Close()
	lib.RabbitMQChan = *ch

	go func() {
		<-conn.NotifyClose(make(chan *amqp.Error))
	}()
	
	/// 메시지 큐 생성
	lib.RabbitMQFirstInit()

	log.Print("RabbitMQ Channel ready")

	log.Print("Redis Connection ready")

	lib.InitFirebase() 
	log.Print("Firebase ready")
	log.Println("Starting the HTTP server on port 50000")

	// <-quit
	// app := routes.InitaliseHandlers()
	// /// TODO : ExchangeDeclare 선언 필요 위치 FanOut이나 direct

	// return app

	// log.Fatal(app.Listen(fmt.Sprintf(":%v", lib.CHAT_SERVER_PORT)))

}
