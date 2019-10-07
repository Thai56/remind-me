package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Message struct {
	Sid         string `json:"sid"`
	Name        string `json:"name"`
	Destination string `json:"destination"`
	Content     string `json:"content"`
	ExpireTime  int64  `json:"expire"`
}

func (a *Api) remind(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	decoder := json.NewDecoder(r.Body)

	var msg Message
	err := decoder.Decode(&msg)

	// output, err := json.Marshal(msg)
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// 	return
	// }

	fmt.Println("MESSAGE:", r.Body, msg.Sid, msg.Content, msg.Destination, msg.ExpireTime)
	err = a.subscriber.QueueReminderFor(msg.Content, msg.Destination, msg.ExpireTime)
	if err != nil {
		fmt.Println(fmt.Errorf("%s", err))
	}

	// w.Write(output)
	p := Payload{
		Data: "success",
	}
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		fmt.Println("failed")
	}
}

type Payload struct {
	Data string `json:"data"`
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	payload := Payload{
		Data: "Hello World",
	}

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		fmt.Println("failed")
	}
}
