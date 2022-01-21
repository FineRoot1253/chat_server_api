package rabbitmq

import (
	"testing"

	"github.com/streadway/amqp"
)

func Test_service_PublishMessage(t *testing.T) {
	type fields struct {
		channel    *amqp.Channel
		repository Repository
	}
	type args struct {
		roomId int
		body   []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &service{
				channel:    tt.fields.channel,
				repository: tt.fields.repository,
			}
			if err := service.PublishMessage(tt.args.roomId, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("service.PublishMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
