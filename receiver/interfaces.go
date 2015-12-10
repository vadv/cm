package receiver

type Storage interface {
	Add(interface{})
}

type Log interface {
	Write(string, string, ...interface{})
}

type ReceiverConfig interface {
	GetLog() interface{}
	GetCommonStorage() interface{}
	GetSettings(string) []byte
}
