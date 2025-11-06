package gen_ids

import (
	"fmt"
	"time"
)

type ObID struct {
	Prefix       string          `json:"prefix"`
	Key          string          `json:"key" bson:"key"`
	LatestId     int64           `json:"latest_id" bson:"latest_id"`
	Date         int             `json:"date"`
	GetIdChannel chan chan int64 `json:"get_id_channel" bson:"get_id_channel"`
	MaxLen       int             `json:"max_len" bson:"max_len"`
}

var ObjIDs map[string]ObID

func InitGenIDservice() {
	// haven't db
	prefixs := []string{"OR", "GP", "", "BG"}

	ObjIDs = map[string]ObID{}

	for _, prefix := range prefixs {
		ObjIDs[prefix] = ObID{
			Prefix:       prefix,
			LatestId:     1,
			Date:         time.Now().Day(),
			GetIdChannel: make(chan chan int64, 1000),
			MaxLen:       9,
		}
	}

	for k, ob := range ObjIDs {
		go func(k string, ob ObID) {
			for {
				select {
				case v, ok := <-ob.GetIdChannel:
					if ok {
						v <- ob.LatestId
						if ob.Date != time.Now().Day() {
							ob.LatestId = 1
						} else {
							ob.LatestId++
						}

					}

				}
			}
		}(k, ob)
	}

}

func GetId(prefix string) string {
	id := make(chan int64, 1)
	ObjIDs[prefix].GetIdChannel <- id

	data := <-id

	gen_id := fmt.Sprint(data)

	if ObjIDs[prefix].MaxLen > len(gen_id) {
		gt := ObjIDs[prefix].MaxLen - len(gen_id)

		for i := 0; i < gt; i++ {
			gen_id = "0" + gen_id
		}

	}
	date := time.Now().Format("20060102")
	if prefix == "" {
		return date + gen_id
	}
	return prefix + date + "-" + gen_id
}

func GetIdOrderId() string {
	return GetId("GP")
}

func GetIdTransactionId() string {
	return GetId("")
}
