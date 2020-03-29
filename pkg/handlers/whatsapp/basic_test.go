package whatsapp

import (
	"reflect"
	"testing"

	"github.com/Rhymen/go-whatsapp"
)

func TestLogin(t *testing.T) {
	type args struct {
		wac      *whatsapp.Conn
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"setup session", args{&whatsapp.Conn{}, "./session.gob"}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Login(tt.args.wac, tt.args.filepath); (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_readSession(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    whatsapp.Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readSession(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("readSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readSession() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeSession(t *testing.T) {
	type args struct {
		session  whatsapp.Session
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeSession(tt.args.session, tt.args.filepath); (err != nil) != tt.wantErr {
				t.Errorf("writeSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageHandler_HandleError(t *testing.T) {
	type fields struct {
		wac       *whatsapp.Conn
		startTime uint64
	}
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh := &MessageHandler{
				wac:       tt.fields.wac,
				startTime: tt.fields.startTime,
			}
			wh.HandleError(tt.args.err)
		})
	}
}

func TestMessageHandler_HandleTextMessage(t *testing.T) {
	type fields struct {
		wac       *whatsapp.Conn
		startTime uint64
	}
	type args struct {
		message whatsapp.TextMessage
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh := &MessageHandler{
				wac:       tt.fields.wac,
				startTime: tt.fields.startTime,
			}
			wh.HandleTextMessage(tt.args.message)
		})
	}
}
