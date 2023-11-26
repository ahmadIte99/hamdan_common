package nats

import (
	"encoding/json"
	"fmt"
	"sync"

	stan "github.com/nats-io/stan.go"
)

type callback func(m *stan.Msg)

var once sync.Once

var Client stan.Conn

func Connect(cluster string, clientId string, url string) stan.Conn {
	// once.Do(func() {
	sc, err := stan.Connect(cluster, clientId, stan.NatsURL(url))
	if err != nil {
		fmt.Println("connect: ", err)
	}
	Client = sc
	// })
	return Client
}

func Publish(topic string, data interface{}) {
	if Client == nil {
		err := fmt.Errorf("publish: there is no connected NATs client")
		fmt.Println(err)
		return
	}
	j, _ := json.Marshal(data)
	Client.Publish(topic, []byte(j))
}

func Listen(topic string, queue string, DurableName string, callback callback) {
	if Client == nil {
		err := fmt.Errorf("listen: there is no connected NATs client")
		fmt.Println(err)
		return
	}
	Client.QueueSubscribe(topic, queue, func(m *stan.Msg) {
		callback(m)
	},
		stan.DeliverAllAvailable(),
		stan.DurableName(DurableName),
		stan.SetManualAckMode(),
	)
}
