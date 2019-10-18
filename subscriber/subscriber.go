package subscriber

import (
	"fmt"
	"strings"
	"log"

	redis "github.com/garyburd/redigo/redis"
)
type ReminderFunc func(string) error

type Subscriber struct {
	mRedisServer string
	mRedisConn   redis.Conn
	remindMethod  ReminderFunc
}

// New - Returns an instance of new Redis Client and Connection.
func New(server string) *Subscriber {
	return &Subscriber{
		mRedisServer: server,
	}
}

func (rc *Subscriber) RegisterReminder(s ReminderFunc) {
	rc.remindMethod = s
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
			// log.Debug("PMessage : pattern: %s \n Channel : %s \n Data : %s\n", msg.Pattern, msg.Channel, msg.Data)

			if chanSlice := strings.Split(msg.Channel, ":"); chanSlice[len(chanSlice)-1] == "expired" {
				messageData := fmt.Sprintf("%s", msg.Data)
				key := fmt.Sprintf("%s", strings.Split(messageData, ":")[len(strings.Split(messageData, ":"))-1])
				
				logMsg := fmt.Sprintf("Key %s Expired", key[len(key)-4:])
				log.Println(logMsg)	
				
				rc.remindMethod(key)
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