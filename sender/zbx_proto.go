package sender

import (
	"encoding/binary"
	"encoding/json"
	"time"
)

var zbxHeader = []byte("ZBXD\x01")

const zbxRequest = `sender data`

type zbxMetric struct {
	Host  string `json:"host"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Clock int64  `json:"clock"`
}

type zbxPacket struct {
	Request string       `json:"request"`
	Data    []*zbxMetric `json:"data"`
	Clock   int64        `json:"clock"`
}

func NewZbxMetric(host, key, value string, clock ...int64) *zbxMetric {
	m := &zbxMetric{Host: host, Key: key, Value: value}
	if m.Clock = time.Now().Unix(); len(clock) > 0 {
		m.Clock = int64(clock[0])
	}
	return m
}

func NewZbxPacket(data []*zbxMetric, clock ...int64) *zbxPacket {
	p := &zbxPacket{Request: zbxRequest, Data: data}
	if p.Clock = time.Now().Unix(); len(clock) > 0 {
		p.Clock = int64(clock[0])
	}
	return p
}

func (packet *zbxPacket) dataLen() []byte {
	dataLen := make([]byte, 8)
	JSONData, _ := json.Marshal(packet)
	binary.LittleEndian.PutUint32(dataLen, uint32(len(JSONData)))
	return dataLen
}

func (packet *zbxPacket) ToBytes() []byte {
	dataPacket, _ := json.Marshal(packet)
	buffer := append(zbxHeader, packet.dataLen()...)
	buffer = append(buffer, dataPacket...)
	return buffer
}
