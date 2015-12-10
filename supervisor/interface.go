package supervisor

type Log interface {
	Write(string, string, ...interface{})
}

type SupervisorConfig interface {
	GetLog() interface{}
	GetSettings(string) []byte
}
