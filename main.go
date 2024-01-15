package main

import (
	"fmt"
	"sync"
	"time"
)

type Message struct {
	Topic   string
	Payload interface{}
}

type Subscriber struct {
	Channel     chan interface{}
	Unsubscribe chan bool
}

type Broker struct {
	subscribers map[string][]*Subscriber
	mutex       sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string][]*Subscriber),
	}
}

func (b *Broker) Subscribe(topic string) *Subscriber {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	subscriber := &Subscriber{
		Channel:     make(chan interface{}, 1),
		Unsubscribe: make(chan bool),
	}

	b.subscribers[topic] = append(b.subscribers[topic], subscriber)

	return subscriber
}

func (b *Broker) Unsubscribe(topic string, subscriber *Subscriber) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if subscribers, found := b.subscribers[topic]; found {

		for i, sub := range subscribers {
			if sub == subscriber {
				close(sub.Channel)
				b.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
				return
			}
		}
	}
}

func (b *Broker) Publish(topic string, payload interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if subscribers, found := b.subscribers[topic]; found {

		for _, sub := range subscribers {
			select {

			case sub.Channel <- payload:
			case <-time.After(time.Second):

				fmt.Printf("Subscriber slow, Ubsubscribing from topic: %s\n", topic)
				b.Unsubscribe(topic, sub)
			}
		}
	}
}

func main() {
	broker := NewBroker()

	subscriber := broker.Subscribe("topic_1")

	go func() {
		for {
			select {
			case msg, ok := <-subscriber.Channel:
				if !ok {
					fmt.Println("Subscriber channel closed")
                    return
				}

				fmt.Printf("Recieved: %v\n", msg)

			case <-subscriber.Unsubscribe:
				fmt.Println("Unsubsribed")
				return
			}
		}
	}()

	broker.Publish("topic_1", "Hello World")
	broker.Publish("topic_1", "Topic 2 sent")
	broker.Publish("topic_1", "Topic 3 sent")
	broker.Publish("topic_1", "Topic 4 sent")
	broker.Publish("topic_1", "Topic 5 sent")
	broker.Publish("topic_1", "Topic 6 sent")

	time.Sleep(2 * time.Second)
	broker.Unsubscribe("topic_1", subscriber)

	broker.Publish("topic_2", "This msg will not be recieved")

	time.Sleep(time.Second)
}
