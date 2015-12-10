package sender

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	defaultZbxAddress          string        = "127.0.0.1:10051"
	defaultZbxQueueSize                      = 256
	defaultZbxMaxQueueSize                   = 16 * 1024
	defaultZbxQueueFlush       time.Duration = 100 * time.Millisecond
	defaultZbxReconnectTimeout time.Duration = 1 * time.Second
	defaultZbxReadTimeout      time.Duration = 2 * time.Second
	defaultZbxWriteTimeout     time.Duration = 2 * time.Second
)

type zbxSender struct {
	sync.Mutex

	Address            string `json:"address"`
	QueueFlushSize     int    `json:"queue_flush_size"`
	QueueMaxSize       int    `json:"queue_max_size"`
	QueueFlushMs       int    `json:"queue_flush_ms"`
	WriteTimeoutMs     int    `json:"write_timeout_ms"`
	ReadTimeoutMs      int    `json:"read_timeout_ms"`
	ReconnectTimeoutMs int    `json:"reconnect_timeout_ms"`

	name string

	log Log

	// очередь
	queue       QueueFifo
	queue_flush time.Duration

	conn               net.Conn
	conn_write_timeout time.Duration
	conn_read_timeout  time.Duration
	conn_reconnect     time.Duration
}

func NewZabbixSender(name string, config SenderConfig) (*zbxSender, error) {
	z := &zbxSender{

		Address:        defaultZbxAddress,
		QueueFlushSize: defaultZbxQueueSize,
		QueueMaxSize:   defaultZbxMaxQueueSize,

		queue_flush: defaultZbxQueueFlush,

		conn_read_timeout:  defaultZbxReadTimeout,
		conn_write_timeout: defaultZbxWriteTimeout,
		conn_reconnect:     defaultZbxReconnectTimeout,

		name: name,

		log:   config.GetLog().(Log),
		queue: config.NewQueueFifo().(QueueFifo),
	}
	if err := json.Unmarshal(config.GetSettings(name), z); err != nil {
		return nil, err
	}
	if z.QueueFlushMs != 0 {
		z.queue_flush = time.Duration(z.QueueFlushMs) * time.Millisecond
	}
	if z.WriteTimeoutMs != 0 {
		z.conn_write_timeout = time.Duration(z.WriteTimeoutMs) * time.Millisecond
	}
	if z.ReadTimeoutMs != 0 {
		z.conn_read_timeout = time.Duration(z.ReadTimeoutMs) * time.Millisecond
	}
	if z.ReconnectTimeoutMs != 0 {
		z.conn_reconnect = time.Duration(z.ReconnectTimeoutMs) * time.Millisecond
	}
	return z, nil
}

func (z *zbxSender) Start() {
	go z.start()
}

func (z *zbxSender) start() {
	flushTicker := time.Tick(z.queue_flush)
	for {
		select {
		case <-flushTicker:
			z.flush()
		}
	}
}

func (z *zbxSender) Inject(event interface{}) {
	go z.inject(event)
}

func (z *zbxSender) inject(e interface{}) {
	event, ok := e.(Event)
	if !ok || event.GetServiceZabbix() == "" {
		return
	}
	z.clearQueue()
	z.queue.Add(event)
	event = nil
	if z.queue.Len() >= z.QueueFlushSize {
		z.flush()
	}
}

// очищаем очередь, если она стала больше чем QueueMaxSize
func (z *zbxSender) clearQueue() {
	if z.queue.Len() < z.QueueMaxSize {
		return
	}
	z.log.Write("ERROR", "[%s] Drop queue (current size: %d, max size: %d)\n", z.name, z.queue.Len(), z.QueueMaxSize)
	for {
		if z.queue.Len() >= z.QueueMaxSize {
			z.queue.Next()
			continue
		}
		break
	}
}

// сбрасываем в con накопленную очередь
func (z *zbxSender) flush() {
	if z.queue.Len() == 0 {
		return
	}
	data := make([]*zbxMetric, 0)
	for {
		if z.queue.Len() == 0 {
			break
		}
		next := z.queue.Next()
		if e, ok := next.(Event); ok {
			metric := NewZbxMetric(e.GetFqdn(), e.GetServiceZabbix(), strconv.FormatFloat(e.GetValue(), 'f', -1, 64))
			data = append(data, metric)
		}
	}
	if len(data) > 0 {
		packet := NewZbxPacket(data)
		z.send(packet.ToBytes())
	}
}

func (z *zbxSender) send(data []byte) {
	z.Lock()
	defer z.Unlock()

	if z.conn != nil {
		z.conn.Close()
	}

	conn, err := net.Dial("tcp", z.Address)
	if err != nil {
		time.Sleep(z.conn_reconnect)
		z.send(data)
	}

	z.conn = conn
	z.conn.SetWriteDeadline(time.Now().Add(z.conn_write_timeout))
	if _, err := z.conn.Write(data); err != nil {
		z.send(data)
	}
}

func (z *zbxSender) read() {
	res := make([]byte, 1024)
	z.conn.SetReadDeadline(time.Now().Add(z.conn_write_timeout))
	res, err := ioutil.ReadAll(z.conn)
	if err != nil {
		z.log.Write("ERROR", "[%s] Read server response: %#v, error: '%s'\n", z.name, string(res), err.Error())
	}
	if err := z.conn.Close(); err == nil {
		z.conn = nil
	}
}
