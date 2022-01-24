package service

import (
	"fmt"
)

type Db_Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Db       string
}

func (config Db_Config) GetConnConfigs() string {
	connConfigs := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s", config.Host, config.Port, config.User, config.Db, config.Password)
	return connConfigs
}
