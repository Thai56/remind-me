package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	maxLength int = 384
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
	Data  string `json:"data"`
	Error *Error `json:"error"`
}

// Error - request error and is a type on payload struct
type Error struct {
	Message string `json:"message"`
	Code    int64  `json:"code"`
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
	case err := <-errChan:
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
			Code:    500,
		}

		p.Error = responseErr
	}

	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *Api) getWiki(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	log.Printf("BODY: %s", body)

	message, err := NewInboundSMS(string(body))
	if err != nil {
		log.Fatalf("Failed to parse inbound sms: %s", err)
		http.Error(w, "can parse query", http.StatusUnprocessableEntity)
		return
	}

	log.Printf("API Handlers - Calling - Search : key %s - from %s - getWiki", message.Body, message.From)

	separatedQuery := strings.Split(message.Body, " ")
	if keyword := strings.ToLower(separatedQuery[0]); keyword != "wiki" {
		errMsg := fmt.Sprintf("Command not found for key %s keyword %s", message.Body, keyword)
		log.Printf(errMsg)
		http.Error(w, errMsg, http.StatusPreconditionFailed)
		err = a.sender.SendMessage(message.From, fmt.Sprintf("I don't recognize that command.\n Did you mean wiki %s?", message.Body))
		if err != nil {
			fmt.Println("Failed To send message when there was no keyword")
		}
		return
	}

	params := map[string]string{
		"action": "opensearch",
		"search": url.QueryEscape(strings.Join(separatedQuery[1:], " ")),
		"format": "json",
		"limit":  "3",
	}

	wikiUrl := "https://en.wikipedia.org/w/api.php?action=opensearch"

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
		log.Println(fmt.Sprintf("Here's the error: %s", errMsg))
		w.WriteHeader(500)
		responseErr := &Error{
			Message: errMsg,
			Code:    500,
		}

		p.Error = responseErr
		smsErrorMsg := fmt.Sprintf("Sorry I Couldn't find anything for %s \n. Please try with a more descriptive search...", message.Body)
		err = a.sender.SendMessage(message.From, smsErrorMsg)
		if err != nil {
			fmt.Println("Failed To send message when failed when wiki failed")
		}
		return
	}

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to read body: %s", err))
		err = a.sender.SendMessage(message.From, "Failed to read response body")
		if err != nil {
			fmt.Println("Failed To send message when failed to read response body")
		}
		return
	}

	results := []interface{}{}
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		log.Println("Failed to unmarshal json: ", err)
		return
	}

	title := results[0].(string)
	otherNames := newUntyped(results[1])
	info := newUntyped(results[2])
	references := newUntyped(results[3])

	msgInfo := strings.Join(info.convertToList(), " ")

	log.Println("MsgInfo ", msgInfo)

	if strings.TrimSpace(msgInfo) == "" {
		fmt.Println("No Message Info Results")
		msgInfo = fmt.Sprintf("Could not find any results for %s. Did you mean %+v?", title, otherNames.convertToList())
	} else if messageLength := len(msgInfo); messageLength > maxLength {
		msgInfo = msgInfo[:maxLength]
		fmt.Println("length of message info %d was greater than %d", messageLength, maxLength)
		fmt.Println("new Message info : %s", msgInfo)
	}

	err = a.sender.SendMessage(message.From, msgInfo)
	if err != nil {
		errMsg := "Failed to send message to sender"
		fmt.Println("API Handlers - %s:%s - getWiki", errMsg, err)
		log.Fatalf("Failed to send")
	}

	fmt.Println("results:", title, otherNames.convertToList(), info.convertToList(), references.convertToList())
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
		log.Printf("Api Handlers - Failed to Parse Query: %v - NewInboundSMS", err)
		return nil, err
	}
	values := map[string]string{}
	for key, _ := range q {
		values[key] = q.Get(key)
	}

	log.Printf("Api Handlers - Values Map %s - NewInboundSMS", values, values["Body"])

	messageBody := values["Body"]
	if len(messageBody) == 0 || strings.TrimSpace(messageBody) == "" {
		errMsg := fmt.Sprintf("Message not found in response from inbound sms.")
		log.Printf("Api Handlers - %s \nvalues: %+v - NewInboundSMS", errMsg, values)
		return nil, errors.New(errMsg)
	}

	res := &InboundSMS{
		Body: messageBody,
	}

	sender := values["From"]
	if sender == "" {
		errMsg := fmt.Sprintf("No source phone number found for inbound sms.")
		log.Printf("Api Handlers - %s \nvalues: %+v - NewInboundSMS", errMsg, values)
		return nil, errors.New(errMsg)
	}

	if res.BodyContainsPlusSeperators() {
		log.Printf("Contained + sign %s", res.Body)
		res.Body = strings.Replace(res.Body, "+", " ", -1)
	}

	res.From = sender
	return res, nil
}

type InboundSMS struct {
	// ToCountry     string `json:"ToCountry"`
	// ToState       string `json:"ToState"`
	// SmsMessageSid string `json:"SmsMessageSid"`
	// NumMedia      string `json:"NumMedia"`
	// ToCity        string `json:"ToCity"`
	// FromCity      string `json:"FromCity"`
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

func (is *InboundSMS) BodyContainsPlusSeperators() bool {
	for _, b := range is.Body {
		if string(b) == "+" {
			return true
		}
	}

	return false
}

func (a *Api) PingIncomingMessage(w http.ResponseWriter, r *http.Request) {
	log.Println("Handlers - Calling - PingIncomingMessage")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	log.Printf("Handlers - Body - %s - PingIncomingMessage", body)
}
