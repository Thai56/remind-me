package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"log"
)

type Message struct {
	Sid          string   `json:"sid"`
	Name         string   `json:"name"`
	Destinations []string `json:"destinations"`
	Content      string   `json:"content"`
	ExpireTime   int64    `json:"expire"`
}

func (a *Api) pingRedis(w http.ResponseWriter, r *http.Request) {
	res, err := a.sender.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Write([]byte(res))
}

func SetKeyWithExpirer() func(destination, message string, expireTime int64) error {
	counter := 0
	return func (destination, message string, expireTime int64) error {
		if counter > 0 {
			fmt.Println("counter", counter)
			return fmt.Errorf("Failed for : %s", destination)
		}
		counter++
		time.Sleep(time.Second * 2)
		return nil
	}
}

func (a *Api) remind(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var err error
	var msg Message
	err = decoder.Decode(&msg)

	fmt.Println("MESSAGE:", r.Body, msg.Sid, msg.Content, msg.Destinations, msg.ExpireTime, len(msg.Destinations))
	
	errChan := make(chan error, len(msg.Destinations))
	defer close(errChan)
	
	var wg sync.WaitGroup
	// SetKeyWithExpire := SetKeyWithExpirer()
	wg.Add(len(msg.Destinations))
	for _, dest := range msg.Destinations {
		go func(d string, wg *sync.WaitGroup) {
			defer wg.Done()
			// err = SetKeyWithExpire(d, msg.Content, msg.ExpireTime)
			err = a.sender.SetKeyWithExpirerer(d, msg.Content, msg.ExpireTime)
			if err != nil {
				log.Println("Handlers - %s - Failed to setKeyWithExpire: %s - remind", dest, err)
				errChan <- err
			}
		}(dest, &wg)
	}

	select {
	case err := <- errChan:
		errMsg := fmt.Sprintf("Handlers - SetKeyMethod Failed for phone number: %s - remind", err)
		log.Println(errMsg)
		return
	default:
		fmt.Println(fmt.Sprintf("Handlers - Finished Saving - %+v - remind", msg.Destinations))
		wg.Wait()
	}

	p := &Payload{
		Data: "success",
	}
	
	if len(errChan) > 0 {
		errMsg := fmt.Sprintf("Could Not Save Reminder")
		log.Println(errMsg)
		w.WriteHeader(500)
		responseErr := &Error{
			Message: errMsg,
			Code: 500,
		}
		if err != nil {
			log.Println("Failed to marshall response data : ", err)
			responseErr.Message = fmt.Sprintf("Failed to marshal response data : %s \n original error: %s", err, responseErr.Message)
		}

		p.Error = responseErr
	}
	
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Fatal(err)
	}
}

type Payload struct {
	Data string `json:"data"`
	Error *Error `json:"error"`
}

type Error struct {
	Message string `json:"message"`
	Code int64 `json:"code"`
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
