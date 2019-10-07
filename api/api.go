package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/user/sms/subscriber"
)

type Api struct {
	name string
	// client     *http.Server
	subscriber *subscriber.Subscriber
}

func New(name string) *Api {
	a := &Api{
		name: name,
		// client: &http.Server{},
	}

	return a
}

func (a *Api) RegisterSubscriber(s *subscriber.Subscriber) {
	a.subscriber = s
}

func (a *Api) Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/Hello", sayHello)
	r.HandleFunc("/", sayHello)
	r.HandleFunc("/remind", a.remind).Methods("POST")

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

		AllowedHeaders: []string{
			"*", //or you can your header key values which you are using in your application

		},
	})

	log.Fatal(http.ListenAndServe(":8080", corsOpts.Handler(r)))
}
