package sender

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	defaultOpenTsdbAddress string        = "127.0.0.1:4242"
	defaultOpenTsdbFlush   time.Duration = 100 * time.Millisecond
	defaultOpenTsdbRetry   time.Duration = 1 * time.Second
	defaultOpenTsdbSize                  = 8 * 1024
)

type opentsdbSender struct {
	sync.Mutex

	Address string `json:"address"`
	RetryMs int    `json:"retry_in_ms"`
	FlushMs int    `json:"flush_in_ms"`
	Size    int    `json:"queue_size_in_b"`

	connected bool
	conn      net.Conn
	writer    *bufio.Writer
	name      string
	retry     time.Duration
	flush     time.Duration
	log       Log
}

func NewOpenTsdbSender(name string, config SenderConfig) (*opentsdbSender, error) {
	s := &opentsdbSender{
		Address: defaultOpenTsdbAddress,
		retry:   defaultOpenTsdbRetry,
		flush:   defaultOpenTsdbFlush,
		Size:    defaultOpenTsdbSize,
		name:    name,
		log:     config.GetLog().(Log),
	}
	if err := json.Unmarshal(config.GetSettings(name), s); err != nil {
		return nil, err
	}
	if s.FlushMs != 0 {
		s.flush = time.Duration(s.FlushMs) * time.Millisecond
	}
	if s.RetryMs != 0 {
		s.flush = time.Duration(s.RetryMs) * time.Millisecond
	}
	return s, nil
}

func (s *opentsdbSender) Start() {
	go s.start()
}

func (s *opentsdbSender) start() {
	s.log.Write("INFO", "[%s] Starting with %s\n", s.name, s.Address)
	flushTicker := time.Tick(s.flush)
	go s.loopConnect()
	for {
		select {
		case <-flushTicker:
			if s.isConnected() {
				if err := s.writer.Flush(); err != nil {
					s.log.Write("ERROR", "[%s] Write error: %s\n", s.name, err.Error())
					s.loopConnect()
				}
			}
		}
	}
}

func (s *opentsdbSender) connect() error {
	s.Lock()
	defer s.Unlock()
	if s.conn != nil {
		s.conn.Close()
	}
	conn, err := net.Dial("tcp", s.Address)
	if err != nil {
		s.log.Write("ERROR", "[%s] Connection error: %s\n", s.name, err.Error())
		return err
	}
	s.log.Write("INFO", "[%s] Connected to: %s\n", s.name, s.Address)
	s.conn = conn
	s.writer = bufio.NewWriterSize(s.conn, s.Size)
	s.connected = true
	return nil
}

func (s *opentsdbSender) isConnected() bool {
	s.Lock()
	defer s.Unlock()
	return s.connected
}

func (s *opentsdbSender) loopConnect() {
	s.Lock()
	s.connected = false
	s.Unlock()
	for {
		if err := s.connect(); err != nil {
			time.Sleep(s.retry)
			continue
		}
		break
	}
}

func (s *opentsdbSender) Inject(event interface{}) {
	go s.inject(event)
}

func (s *opentsdbSender) inject(e interface{}) {
	event, ok := e.(Event)
	if !ok || !s.isConnected() || event.GetServiceTSDB() == "" {
		return
	}
	data := s.convertEvent(event)
	if _, err := s.writer.Write(data); err != nil {
		s.log.Write("ERROR", "[%s] Write error: %s\n", s.name, err.Error())
		s.loopConnect()
	}
}

func (s *opentsdbSender) convertEvent(e Event) []byte {
	var tags string
	for key, val := range e.GetTags() {
		tags = fmt.Sprintf("%s %s=%s", tags, key, val)
	}
	return []byte(fmt.Sprintf("put %s %v %v %s\n", e.GetServiceTSDB(), e.GetTime(), e.GetValue(), tags))
}
