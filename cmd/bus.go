package main

import (
	"cm/event"
	"cm/log"
	"cm/storage"
)

type Reciever interface {
	Start()
}

type Sender interface {
	Start()
	Inject(interface{})
}

type SupervisorTask interface {
	Start()
}

type BusConfig struct {
	recievers       map[string]Reciever
	senders         map[string]Sender
	supervisorTasks map[string]SupervisorTask
	log             *log.Logger
	settings        map[string][]byte
	commonStorage   *storage.QueueFifo
}

func newBusConfig(log *log.Logger) *BusConfig {
	return &BusConfig{
		log:             log,
		settings:        make(map[string][]byte),
		recievers:       make(map[string]Reciever),
		senders:         make(map[string]Sender),
		supervisorTasks: make(map[string]SupervisorTask),
	}
}

func (b *BusConfig) setSettings(name string, config []byte) {
	b.settings[name] = config
}

func (b *BusConfig) addNewSender(name string, sender Sender) {
	b.senders[name] = sender
}

func (b *BusConfig) addNewReciever(name string, reciever Reciever) {
	b.recievers[name] = reciever
}

func (b *BusConfig) addSupervisorTasks(name string, t SupervisorTask) {
	b.supervisorTasks[name] = t
}

func (b *BusConfig) setCommonStorage(s *storage.QueueFifo) {
	b.commonStorage = s
}

func (b *BusConfig) GetLog() interface{} {
	return b.log
}

func (b *BusConfig) NewQueueFifo() interface{} {
	return storage.NewQueueFifo()
}

func (b *BusConfig) GetSettings(name string) []byte {
	return b.settings[name]
}

func (b *BusConfig) GetCommonStorage() interface{} {
	return b.commonStorage
}

func (bus *BusConfig) Flush() {
	for {
		if e := bus.commonStorage.Next(); e != nil {
			if events, err := event.Parse(e.([]byte)); err == nil {
				for _, event := range events {
					for _, sender := range bus.senders {
						sender.Inject(event)
					}
				}
			}
		} else {
			break
		}
	}
}
