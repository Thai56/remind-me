package main

import (
	"github.com/user/sms/api"
	"github.com/user/sms/sender"
	"github.com/user/sms/subscriber"
	sub "github.com/user/sms/subscriber"
)

func main() {
	app := New("sms")
	app.subscriber.RegisterReminder(app.sender.HandleReminder)
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
	sender     *sender.Sender
}

func New(name string) *App {
	return &App{
		name:       name,
		subscriber: subscriber.New("_subscriber_server"),
		sender:     sender.New("sms_sender"),
	}
}
