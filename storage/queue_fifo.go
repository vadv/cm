package storage

import (
	"sync"
)

const chunkFifoSize = 512

type chunkFifo struct {
	items       [chunkFifoSize]interface{}
	first, last int
	next        *chunkFifo
}

type QueueFifo struct {
	sync.Mutex
	head, tail *chunkFifo // chunk head and tail
	count      int
}

func NewQueueFifo() (q *QueueFifo) {
	initChunkFifo := new(chunkFifo)
	q = &QueueFifo{
		head: initChunkFifo,
		tail: initChunkFifo,
	}
	return q
}

func (q *QueueFifo) Len() (length int) {
	q.Lock()
	defer q.Unlock()
	length = q.count
	return length
}

func (q *QueueFifo) Add(item interface{}) {
	q.Lock()
	defer q.Unlock()
	if q.tail.last >= chunkFifoSize {
		q.tail.next = new(chunkFifo)
		q.tail = q.tail.next
	}
	q.tail.items[q.tail.last] = item
	q.tail.last++
	q.count++
}

func (q *QueueFifo) Next() (item interface{}) {
	q.Lock()
	defer q.Unlock()

	if q.count == 0 {
		return nil
	}

	if q.head.first >= q.head.last {
		return nil
	}

	item = q.head.items[q.head.first]

	q.head.first++
	q.count--

	if q.head.first >= q.head.last {
		if q.count == 0 {
			q.head.first = 0
			q.head.last = 0
			q.head.next = nil
		} else {
			q.head = q.head.next
		}
	}

	return item
}
