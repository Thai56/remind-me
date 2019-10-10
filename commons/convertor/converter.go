package convert

import (
	"time"
	"strconv"
)

func Timestamp() string {
	t := time.Now()
	return t.Format(time.RFC3339)
}

func UnixToTimestamp(unixTime string) string {
	i, err := strconv.ParseInt("1405544146", 10, 64)
    if err != nil {
        panic(err)
    }
	tm := time.Unix(i, 0).String()
	
	return tm
}