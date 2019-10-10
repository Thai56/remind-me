package main

import (
	"github.com/user/sms/api"
	"github.com/user/sms/sender"
	"github.com/user/sms/storage"
	store "github.com/user/sms/storage"
	"github.com/user/sms/subscriber"
	sub "github.com/user/sms/subscriber"
)

func main() {
	app := New("sms")
	app.store.InitUserData()
	app.subscriber.RegisterSender(app.sender)
	smsAPI := api.New("sms")
	smsAPI.RegisterSubscriber(app.subscriber)
	smsAPI.RegisterSender(app.sender)
	go smsAPI.Serve()
	app.subscriber.Run()
	return
}

type App struct {
	name       string
	subscriber *sub.Subscriber
	store      *store.Storage
	sender     *sender.Sender
}

func New(name string) *App {
	return &App{
		name:       name,
		subscriber: subscriber.New("_subscriber_server"),
		store:      storage.New("sms_storage"),
		sender:     sender.New("sms_sender"),
	}
}
