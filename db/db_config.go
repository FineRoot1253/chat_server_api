package db

import "fmt"

type Db_Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Db       string
}

var GetConnConfigs = func(config Db_Config) string {
	connConfigs := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s", config.Host, config.Port, config.User, config.Db, config.Password)
	return connConfigs
}
