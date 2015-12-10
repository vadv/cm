package sender

type SenderConfig interface {
	GetLog() interface{}
	GetSettings(string) []byte
	NewQueueFifo() interface{}
}

type Log interface {
	Write(string, string, ...interface{})
}

type Event interface {
	GetFqdn() string
	GetService() string
	GetServiceTSDB() string
	GetServiceGraphite() string
	GetServiceZabbix() string
	GetValue() float64
	GetTime() int64
	GetTags() map[string]string
}

type QueueFifo interface {
	Add(interface{})
	Next() interface{}
	Len() int
}
