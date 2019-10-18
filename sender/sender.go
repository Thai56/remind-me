package sender

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
	"errors"

	redis "github.com/garyburd/redigo/redis"
	uuid "github.com/satori/go.uuid"
	"github.com/user/sms/config"
)

type Sender struct {
	name        string
	pool *redis.Pool
}

func (s *Sender) Ping() (string, error) {
	conn := s.pool.Get()
	defer conn.Close()
	return redis.String(conn.Do("PING"))
}

func New(name string) *Sender {
	pool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379")
		},
	}

	return &Sender{
		name:        name,
		pool: pool,
	}
}

// SendMessage -
func (s *Sender) SendMessage(destination, message string) error {
	rand.Seed(time.Now().Unix())

	msgData := url.Values{}
	msgData.Set("To", destination)
	msgData.Set("From", config.GetAssignedPhoneNumber())
	msgData.Set("Body", message)
	msgDataReader := *strings.NewReader(msgData.Encode())

	err := SendRequest(&msgDataReader)
	if err != nil {
		fmt.Println("FAILED TO SEND REQUEST ", err)
	}

	return err
}

// SendRequest -
func SendRequest(body *strings.Reader) error {
	accountSid := config.GetAccountSid()
	authToken := config.GetAuthToken()
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + accountSid + "/Messages.json"
	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, body)
	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println("sid data", data["sid"])
		}
	} else {
		fmt.Println("status", resp.Status)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("body", string(body))
	}

	return err
}

type Error struct {
	message string
	code int64
}

type MetaData struct {
	Message     string
	Destination string
}

func (s *Sender) SetKeyWithExpirerer(destination, message string, expireTime int64) error {
	fmt.Println("Sender - Calling - SetKeyWithExpirerer Method ", destination)
	
	conn := s.pool.Get()
	defer conn.Close()
	
	// generate unique key to store the meta data.
	uuID := fmt.Sprintf("%x", uuid.Must(uuid.NewV4()))
	shadowKey := fmt.Sprintf("shadowkey:%s", uuID)
	conn.Send("MULTI")
	conn.Send("Set", fmt.Sprintf("%s:%s", uuID, destination), message)
	conn.Send("Set", shadowKey, time.Now().Unix())
	conn.Send("PEXPIREAT", shadowKey, expireTime)
	_, err := conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// GetMetaDataFor - finds all data for the key given and will return the result at zero index.
func (s *Sender) GetMetaDataFor(key string) (*MetaData, error) {
	var errMsg string
	var method string = "GetMetaDataFor" 

	conn := s.pool.Get()
	defer conn.Close()
	
	matches, err := redis.ByteSlices(conn.Do("KEYS", fmt.Sprintf("%s:*", key)))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to find matching key for : %s : %s", key , err)
		logFrom(method, errMsg)
		return nil, fmt.Errorf("Failed find keys for %s : %s", key, err)
	}
	
	if len(matches) == 0 {
		errMsg = fmt.Sprintf("No matches found for key : %s : %s",key , err)
		logFrom(method, errMsg)
		return nil, errors.New(errMsg)
	}
	
	keyWithDestination := fmt.Sprintf("%s", matches[0])
	destination := strings.Split(keyWithDestination, ":")[1]
	
	message, err := redis.String(conn.Do("Get", keyWithDestination))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to get message for key: %s : %s", key, err)
		logFrom(method, errMsg)
		return nil, errors.New(errMsg)
	}

	return &MetaData{
		Message:     message,
		Destination: destination,
		}, nil
}

func (s *Sender) RemoveAllInstancesOf(key string) error {
	var errMsg string

	conn := s.pool.Get()
	defer conn.Close()

	method := "RemoveAllInstancesOf"
	matches, err := redis.ByteSlices(conn.Do("KEYS", fmt.Sprintf("%s:*", key)))
	if err != nil {
		errMsg = fmt.Sprintf("Failed find keys for %s : %s", key, err)
		logFrom(method, errMsg)
		return errors.New(errMsg)
	}
	
	for _, b := range matches {
		str := fmt.Sprintf("%s", b)
		fmt.Println(str)
		_, err := conn.Do("DEL", str)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to delete key : %s : %s", key, err)
			logFrom(method, errMsg)
		}
	}

	return err
}

func (s *Sender) HandleReminder(key string) error {
	var err error
	var errMsg string

	// Logging variables
	method := "HandleReminder"
	shortKey := key[:len(key)-4]
	
	metadata, err := s.GetMetaDataFor(key)
	if err != nil {
		errMsg := fmt.Sprintf("Sender - Failed to get metadata for key: %s - %s", shortKey, err)
		logFrom(method, errMsg)
		return errors.New(errMsg)
	}

	if metadata.Destination == "" {
		errMsg = fmt.Sprintf("No Destination found for key : %s", shortKey)
		logFrom(method, errMsg)
		return errors.New(errMsg)
	}

	logFrom(method, fmt.Sprintf("sending message : %s", metadata))
	err = s.SendMessage(metadata.Destination, metadata.Message)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to send message for %s : %s", metadata.Destination, err)
		logFrom(method, errMsg)
		return errors.New(errMsg)
	}

	err = s.RemoveAllInstancesOf(key)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to delete key for key : %s : %s", shortKey, err)
		logFrom(method, errMsg)
		return errors.New(errMsg)
	}

	return nil
}

// ====== //
// Logger //
// ====== //

func logFrom(method, msg string) {
	log.Printf("Sender - %s - %s", msg, method)
}