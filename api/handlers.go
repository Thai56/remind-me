package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"log"
	"net/url"
	"io/ioutil"
)

type Message struct {
	Sid          string   `json:"sid"`
	Name         string   `json:"name"`
	Destinations []string `json:"destinations"`
	Content      string   `json:"content"`
	ExpireTime   int64    `json:"expire"`
}

// Payload - request return type
type Payload struct {
	Data string `json:"data"`
	Error *Error `json:"error"`
}

// Error - request error and is a type on payload struct
type Error struct {
	Message string `json:"message"`
	Code int64 `json:"code"`
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
	
func (a *Api) setReminder(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var err error
	var msg Message
	err = decoder.Decode(&msg)
	
	fmt.Println("MESSAGE:", r.Body, msg.Sid, msg.Content, msg.Destinations, msg.ExpireTime, len(msg.Destinations))
	
	errChan := make(chan error, len(msg.Destinations))
	defer close(errChan)
	
	var wg sync.WaitGroup
	wg.Add(len(msg.Destinations))
	for _, dest := range msg.Destinations {
		go func(d string, wg *sync.WaitGroup) {
			defer wg.Done()
			err = a.sender.SetKeyWithExpirerer(d, msg.Content, msg.ExpireTime)
			if err != nil {
				log.Println("Handlers - %s - Failed to setKeyWithExpire: %s - setReminder", dest, err)
				errChan <- err
			}
			}(dest, &wg)
		}
		
		select {
		case err := <- errChan:
			errMsg := fmt.Sprintf("Handlers - SetKeyMethod Failed for phone number: %s - setReminder", err)
		log.Println(errMsg)
	default:
		fmt.Println(fmt.Sprintf("Handlers - Finished Saving - %+v - setReminder", msg.Destinations))
		wg.Wait()
	}
	
	p := &Payload{
		Data: "success",
	}
	
	if len(errChan) > 0 {
		errMsg := fmt.Sprintf("Could Not Save Reminder %+v", msg.Destinations)
		log.Println(errMsg)
		w.WriteHeader(500)
		responseErr := &Error{
			Message: errMsg,
			Code: 500,
		}

		p.Error = responseErr
	}
	
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *Api) getWiki(w http.ResponseWriter, r *http.Request) {
	// keys, ok := r.URL.Query()["search"]
	// if !ok || len(keys[0]) < 1 {
	// 	log.Println("Url Param 'key' is missing")
	// 	// TODO: Error
    //     return
    // }
	
	// key := keys[0]

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	log.Printf("BODY: %s", body)

	message, err := NewInboundSMS(string(body))
	if err != nil {
		http.Error(w, "can parse query", http.StatusUnprocessableEntity)
		return
	}

	log.Printf("%+v", message)
	
	if message.Body == "" {
		log.Fatal("No Body found in request")
	}
	
	log.Printf("Handlers - Key: %s - getWiki", message.Body)
		
	logMsg := fmt.Sprintf("API Handlers - Calling - Search : %s - getWiki", message.Body)
	log.Println(logMsg)
	
	params := map[string]string{
		"action": "opensearch",
		"search": url.QueryEscape(message.Body),
		"format": "json",
		"limit": "3",
	};
	wikiUrl := "https://en.wikipedia.org/w/api.php?action=opensearch" // &search=Nelson%20Mandela&format=json&limit=5"

	for k, v := range params {
		wikiUrl += fmt.Sprintf("&%s=%s", k, v)
	}
	log.Println("wikiUrl ", wikiUrl)
	
	p := &Payload{
		Data: "success",
	}

	response, err := http.Get(wikiUrl)
	if err != nil {
		errMsg := fmt.Sprintf("API Handlers - Could Not Get Wiki for url: %s - GetWiki", err)
		log.Println(errMsg)
		w.WriteHeader(500)
		responseErr := &Error{
			Message: errMsg,
			Code: 500,
		}

		p.Error = responseErr
	}
	
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	
	results := []interface{}{}
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		log.Fatal("Failed to unmarshal json: ", err)
	}

	title := results[0].(string)
	otherNames := newUntyped(results[1])
	info := newUntyped(results[2])
	references := newUntyped(results[3])
	
	fmt.Println("results:", title, otherNames.convertToList(), info.convertToList(), references.convertToList()[:2])
}

func newUntyped(t interface{}) Untyped {
	return Untyped{
		t: t,
	}
}
type Untyped struct {
	t interface{}
}

func (u *Untyped) convertToList() []string {
	result := []string{}
	for _, val := range u.t.([]interface{}) {
		result = append(result, val.(string))
	}
	return result
}


func NewInboundSMS(body string) (*InboundSMS, error) {
	q, err := url.ParseQuery(body)
	if err != nil {
		log.Printf("Failed to Parse Query: %v", err)
		return nil, err
	}
	values := map[string]string{}
	for key, _ := range q {
		values[key] = q.Get(key)
	}

	js, err := json.Marshal(values)
	if err != nil {
	    panic(err)
	}
	log.Printf("Marshalled %s", js)
	e := &InboundSMS{}
	err = json.Unmarshal(js, e) 
	if err != nil {
		panic(err)
	}
	log.Printf("RETURNING: %+v", e)
	return e, nil
}

type InboundSMS struct {
	ToCountry string `json:"ToCountry"`
	ToState string `json:"ToState"`
	SmsMessageSid string `json:"SmsMessageSid"`
	NumMedia string `json:"NumMedia"`
	ToCity string `json:"ToCity"`
	FromCity string `json:"FromCity"`
	Body string `json:"Body"`
	From string `json:"From"`
	// ToCountry=US
	// ToState=WI
	// SmsMessageSid=SM581f52e533511ac147ae7c4d9a4c9d89
	// NumMedia=0
	// ToCity=LA+CROSSE
	// FromZip=94535
	// SmsSid=SM581f52e533511ac147ae7c4d9a4c9d89
	// FromState=CA
	// SmsStatus=received
	// FromCity=FAIRFIELD
	// Body=Wiki+Donald+trump
	// FromCountry=US
	// To=%2B16084332365
	// ToZip=54650
	// NumSegments=1
	// MessageSid=SM581f52e533511ac147ae7c4d9a4c9d89
	// AccountSid=ACf8513108c1afe25b2cf5616f8d8ff8fb
	// From=%2B17073447433
	// ApiVersion=2010-04-01
}



func(a *Api) PingIncomingMessage(w http.ResponseWriter, r *http.Request) {
	log.Println("Handlers - Calling - PingIncomingMessage")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	log.Printf("Handlers - Body - %s - PingIncomingMessage", body)
}
