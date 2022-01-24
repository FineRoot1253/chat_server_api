package room

import (
	"log"
	"testing"

	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/config"
	srv "github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/service"
)

func Test_service_GetRoomListOfUser(t *testing.T) {

	type args struct {
		key int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "[Success Test]",
			args:    args{key: 1},
			wantErr: false,
		},
		{
			name:    "[Failure Test]",
			args:    args{key: 0},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := srv.Connect(initDB())
			if err != nil {
				t.Errorf("service.GetRoomListOfUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			s := NewService(NewRepository(conn))
			_, err1 := s.GetRoomListOfUser(tt.args.key)
			if (err1 != nil) != tt.wantErr {
				t.Errorf("service.GetRoomListOfUser() error = %v, wantErr %v", err1, tt.wantErr)
				return
			}
		})
	}
}

func initDB() string {
	config :=
		srv.Db_Config{
			Host:     config.HOST,
			Port:     config.POSTGRES_PORT,
			User:     config.POSTGRES_USER,
			Password: config.POSTGRES_PWD,
			Db:       config.POSTGRES_DB,
		}
	log.Print("testsetsetstesetsetseet", config.Host)

	return config.GetConnConfigs()

}
