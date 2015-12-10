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
	defaultGraphiteAddress string        = "127.0.0.1:2003"
	defaultGraphiteFlush   time.Duration = 100 * time.Millisecond
	defaultGraphiteRetry   time.Duration = 1 * time.Second
	defaultGraphiteSize                  = 8 * 1024
	defaultGraphiteNet     string        = "tcp"
)

type graphiteSender struct {
	sync.Mutex
	// json
	Address string `json:"address"`
	Network string `json:"network"`
	RetryMs int    `json:"retry_ms"`
	FlushMs int    `json:"flush_ms"`
	Size    int    `json:"queue_size_b"`

	// real
	retry time.Duration
	flush time.Duration

	conn      net.Conn
	writer    *bufio.Writer
	connected bool
	name      string
	log       Log
}

func NewGrahiteSender(name string, config SenderConfig) (*graphiteSender, error) {

	s := &graphiteSender{
		Address: defaultGraphiteAddress,
		retry:   defaultGraphiteRetry,
		flush:   defaultGraphiteFlush,
		Size:    defaultGraphiteSize,
		Network: defaultGraphiteNet,
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

func (g *graphiteSender) Start() {
	go g.start()
}

func (g *graphiteSender) start() {
	g.log.Write("INFO", "[%s] Starting with %s\n", g.name, g.Address)
	flushTicker := time.Tick(g.flush)
	go g.loopConnect()
	for {
		select {
		case <-flushTicker:
			if g.isConnected() {
				if err := g.writer.Flush(); err != nil {
					g.log.Write("ERROR", "[%s] Write error: %s\n", g.name, err.Error())
					g.loopConnect()
				}
			}
		}
	}
}

func (g *graphiteSender) connect() error {
	g.Lock()
	defer g.Unlock()
	if g.conn != nil {
		g.conn.Close()
	}
	conn, err := net.Dial(g.Network, g.Address)
	if err != nil {
		g.log.Write("ERROR", "[%s] Connection error: %s\n", g.name, err.Error())
		return err
	}
	g.log.Write("INFO", "[%s] Connected to: %s\n", g.name, g.Address)
	g.conn = conn
	g.writer = bufio.NewWriterSize(g.conn, g.Size)
	g.connected = true
	return nil
}

func (g *graphiteSender) isConnected() bool {
	g.Lock()
	defer g.Unlock()
	return g.connected
}

func (g *graphiteSender) loopConnect() {
	g.Lock()
	g.connected = false
	g.Unlock()
	for {
		if err := g.connect(); err != nil {
			time.Sleep(g.retry)
			continue
		}
		break
	}
}

func (g *graphiteSender) Inject(event interface{}) {
	go g.inject(event)
}

func (g *graphiteSender) inject(e interface{}) {
	event, ok := e.(Event)
	if !ok || !g.isConnected() || event.GetServiceGraphite() == "" {
		return
	}
	data := []byte(fmt.Sprintf("%s %f %d\n", event.GetServiceGraphite(), event.GetValue(), event.GetTime()))
	if _, err := g.writer.Write(data); err != nil {
		g.log.Write("ERROR", "[%s] Write error: %s\n", g.name, err.Error())
		g.loopConnect()
	}
}
