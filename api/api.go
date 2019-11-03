package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/user/sms/subscriber"
	"github.com/user/sms/sender"
)

type Api struct {
	name string
	subscriber *subscriber.Subscriber
	sender *sender.Sender
}

func New(name string) *Api {
	a := &Api{
		name: name,
	}

	return a
}

func (a *Api) RegisterSubscriber(s *subscriber.Subscriber) {
	a.subscriber = s
}

func  (a *Api) RegisterSender(s *sender.Sender) {
	a.sender = s
}

func (a *Api) Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/remind", a.setReminder).Methods("POST")
	r.HandleFunc("/ping", a.pingRedis).Methods("GET")
	r.HandleFunc("/wiki", a.getWiki).Methods("GET")
	r.HandleFunc("/rest", a.PingIncomingMessage).Methods("POST")

	corsOpts := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, //you service is available and allowed for this base url
		AllowedMethods: []string{
			http.MethodGet, //http methods for your app
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},

		AllowedHeaders: []string{"*"},
	})

	log.Fatal(http.ListenAndServe(":8080", corsOpts.Handler(r)))
}
