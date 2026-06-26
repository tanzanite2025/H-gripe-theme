package eventbus

import (
	"sync"
)

type EventHandler func(event interface{})

type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

var instance *EventBus
var once sync.Once

func GetBus() *EventBus {
	once.Do(func() {
		instance = &EventBus{
			handlers: make(map[string][]EventHandler),
		}
	})
	return instance
}

func (b *EventBus) Subscribe(topic string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = append(b.handlers[topic], handler)
}

func (b *EventBus) Publish(topic string, event interface{}) {
	b.mu.RLock()
	handlers := b.handlers[topic]
	b.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}

func Subscribe(topic string, handler EventHandler) {
	GetBus().Subscribe(topic, handler)
}

func Publish(topic string, event interface{}) {
	GetBus().Publish(topic, event)
}
