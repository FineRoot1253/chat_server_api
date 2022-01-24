package service

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Connector *gorm.DB

func Connect(configStr string) (*gorm.DB, error) {

	var err error
	var connecter *gorm.DB
	/// postgresql이 완전히 켜질때까지 대기한다.
	for i := 0; i < 10; i++ {
		connecter, err = gorm.Open(postgres.Open(configStr), &gorm.Config{})
		if err != nil {
			log.Printf("Unable to Open DB: %s... Retrying\n", err.Error())
			time.Sleep(time.Second * 2)

		}
		db, err1 := connecter.DB()
		if err := db.Ping(); err != nil || err1 != nil {
			log.Printf("Unable to Ping DB: %s... Retrying\n", err.Error())
			time.Sleep(time.Second * 2)
		} else {
			err = nil
			break
		}

	}

	log.Println("Conn Successed")
	return connecter, nil
}

//Migrate create/updates database table
func Migrate(table *interface{}) {
	Connector.AutoMigrate(&table)
	log.Println("Table migrated")
}
