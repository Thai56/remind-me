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

type WikiParams struct {
	Action string `json:"action"`
	Search string `json:"search"`
	Limit string `json:"limit"`
	Namespace string `json:"namespace"`
	Format string `json:"format"`
}

func (a *Api) getWiki(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["openSearch"]	
	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		// TODO: Error
        return
    }
	
	key := keys[0]
	
	logMsg := fmt.Sprintf("API Handlers - Calling - Search : %s - getWiki", key)
	log.Println(logMsg)
	
	params := map[string]string{
		"action": "opensearch",
		"search": url.QueryEscape(key),
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
	
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	
	results := []interface{}{}
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		log.Fatal("Failed to unmarshal json: ", err)
	}

	title := results[0].(string)
	otherNames := UntypedStringList{ untyped: results[1] }
	info := UntypedStringList{ untyped: results[2] }
	references := UntypedStringList{ untyped: results[3] }
	
	fmt.Println("results:", title, otherNames.convertToList(), info.convertToList(), references.convertToList()[:2])
}

type UntypedStringList struct {
	untyped interface{}
}

func (u *UntypedStringList) convertToList() []string {
	result := []string{}
	for _, val := range u.untyped.([]interface{}) {
		result = append(result, val.(string))
	}
	return result
}