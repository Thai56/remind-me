package subscriber

import (
	"fmt"
	"strings"

	redis "github.com/garyburd/redigo/redis"
	"github.com/user/sms/sender"
	convert "github.com/user/sms/commons/convertor"
)

type Subscriber struct {
	mRedisServer string
	mRedisConn   redis.Conn
	sender       *sender.Sender
}

// New - Returns an instance of new Redis Client and Connection.
func New(server string) *Subscriber {
	return &Subscriber{
		mRedisServer: server,
	}
}

func (rc *Subscriber) RegisterSender(s *sender.Sender) {
	rc.sender = s
}

// Run - Connects to redis then enables "notify-keyspace-events".
// Subscribes to "__key*__:*"
func (rc *Subscriber) Run() {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	rc.mRedisConn = conn
	fmt.Println(conn)
	rc.mRedisConn.Do("CONFIG", "SET", "notify-keyspace-events", "KEA")

	defer rc.mRedisConn.Close()

	psc := redis.PubSubConn{Conn: rc.mRedisConn}
	psc.PSubscribe("__key*__:*")

	for {
		switch msg := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("Message: %s %s\n", msg.Channel, msg.Data)
		case redis.PMessage:
			// fmt.Printf("PMessage : pattern: %s \n Channel : %s \n Data : %s\n", msg.Pattern, msg.Channel, msg.Data)

			if chanSlice := strings.Split(msg.Channel, ":"); chanSlice[len(chanSlice)-1] == "expired" {
				key := getKeySuffixFrom(msg.Data)
				
				logMsg := fmt.Sprintf("Key %s Expired at %s", key[len(key)-4:], convert.Timestamp())
				fmt.Println(logMsg)
				
				metadata, err := rc.sender.GetMetaDataFor(key)
				if err != nil {
					fmt.Println(fmt.Sprintf("%s", err))
					return
				}

				if metadata.Destination == "" {
					fmt.Println("No Destination found for key : ", key)
				}

				fmt.Println("sending message", metadata)
				err = rc.sender.SendMessage(metadata.Destination, metadata.Message)
				if err != nil {
					fmt.Println("Failed to send message for %s", metadata.Destination)
					return
				}

				err = rc.sender.RemoveAllInstancesOf(key)
				if err != nil {
					fmt.Println("Failed to delete key for %s : %s", key, err)
					return
				}
			}
		case redis.Subscription:
			fmt.Printf("Subscription: %s %s %d\n", msg.Kind, msg.Channel, msg.Count)
			if msg.Count == 0 {
				return
			}
		case error:
			fmt.Printf("error: %v\n", msg)
			return
		}
	}
}

func getKeySuffixFrom(msg []byte) string {
	messageData := fmt.Sprintf("%s", msg)
	return fmt.Sprintf("%s", strings.Split(messageData, ":")[len(strings.Split(messageData, ":"))-1])
}

// QueueReminderFor - creates an expire key <user_sid:phone_number:message_hash>
// stores the message under the hash key
func (rc *Subscriber) QueueReminderFor(message, destination string, expireTime int64) error {
	err := rc.sender.SetKeyWithExpire(destination, message, expireTime)
	if err != nil {
		return fmt.Errorf("Failed to set Key: %s", err)
	}

	return nil
}
