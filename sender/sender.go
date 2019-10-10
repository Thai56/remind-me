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

	redis "github.com/garyburd/redigo/redis"
	uuid "github.com/satori/go.uuid"
	"github.com/user/sms/config"
)

type Sender struct {
	name        string
	redisClient redis.Conn
}

func (s *Sender) Ping() (string, error) {
	return redis.String(s.redisClient.Do("PING"))
}

func New(name string) *Sender {
	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Fatal(err)
	}

	return &Sender{
		name:        name,
		redisClient: c,
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

type MetaData struct {
	Message     string
	Destination string
}

// SetKeyWithExpire - stores a key with an expire time.
func (s *Sender) SetKeyWithExpire(destination, message string, expireTime int64) error {
	// generate unique key to store the meta data.
	uuID := fmt.Sprintf("%x", uuid.Must(uuid.NewV4()))

	_, err := s.redisClient.Do("Set", fmt.Sprintf("%s:%s", uuID, destination), message)
	if err != nil {
		return fmt.Errorf("Failed To set the key : %s :%s", uuID, err)
	}

	shadowKey := fmt.Sprintf("shadowkey:%s", uuID)
	_, err = s.redisClient.Do("Set", shadowKey, expireTime)
	if err != nil {
		return fmt.Errorf("Failed To set the key : %s :%s : ", uuID, err)

	}

	_, err = s.redisClient.Do("PEXPIREAT", shadowKey, expireTime)
	if err != nil {
		return fmt.Errorf("Failed To expire the key : %s :%s : ", uuID, err)
	}

	// logMsg := fmt.Sprintf("Set key %s to expire at %s", shadowKey[len(shadowKey)-4:], convert.UnixToTimestamp(fmt.Sprintf("%d", expireTime)))


	return nil
}

// GetMetaDataFor - finds all data for the key given and will return the result at zero index.
func (s *Sender) GetMetaDataFor(key string) (*MetaData, error) {
	matches, err := redis.ByteSlices(s.redisClient.Do("KEYS", fmt.Sprintf("%s:*", key)))
	if err != nil {
		return nil, fmt.Errorf("Failed find keys for %s : %s", key, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("No matches found for key: %s", key)
	}

	keyWithDestination := fmt.Sprintf("%s", matches[0])
	destination := strings.Split(keyWithDestination, ":")[1]

	message, err := redis.String(s.redisClient.Do("Get", keyWithDestination))
	if err != nil {
		return nil, fmt.Errorf("Failed to get message for key: %s : %s", key, err)
	}

	return &MetaData{
		Message:     message,
		Destination: destination,
	}, err

}

func (s *Sender) KeyExists(key string) (bool, error) {
	fmt.Println("Checking if key exists", key)
	exists, err := redis.Bool(s.redisClient.Do("EXISTS", key))

	return exists, err
}

func (s *Sender) RemoveAllInstancesOf(key string) error {
	matches, err := redis.ByteSlices(s.redisClient.Do("KEYS", fmt.Sprintf("%s:*", key)))
	if err != nil {
		return fmt.Errorf("Failed find keys for %s : %s", key, err)
	}

	for _, b := range matches {
		str := fmt.Sprintf("%s", b)
		fmt.Println(str)
		_, err := s.redisClient.Do("DEL", str)
		if err != nil {
			fmt.Println("Failed to delete key ", key, " ", err)
		}
	}
	return err
}
