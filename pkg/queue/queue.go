package queue

import (
	"sync"
	"time"
)

// queue is a struct for queue
type queue struct {
	enqueue chan interface{}
	dequeue chan interface{}
	close   chan bool
	once    sync.Once
}

// Queue is an interface for a queue structure
type Queue interface {

	// Push pushes the value to the end of queue
	Push(value interface{})

	// Poll polls the top most value from the queue
	// If the queue is empty, it will block until a value is available
	Poll() (value interface{}, ok bool)

	// Close closes the queue for write operations
	// if the timeOut is less than 0, it will close the channel to enqueue and keep the queue read only
	Close(timeOut time.Duration)
}

// manage is a function to manage the queue
func (q *queue) manage() {
	queue := []interface{}{}
	defer close(q.dequeue)

	// An infinite loop to periodically check the queue
	for {
		if len(queue) == 0 {
			select {
			case <-q.close:
				return
			case v, ok := <-q.enqueue:
				if !ok {
					return
				}
				queue = append(queue, v)
			}
		} else {
			select {
			case <-q.close:
				return
			case v, ok := <-q.enqueue:
				if ok {
					queue = append(queue, v)
				}
			case q.dequeue <- queue[0]:
				queue[0] = nil
				queue = queue[1:]
			}
		}
	}
}

func (q *queue) Push(value interface{}) {
	q.enqueue <- value
}

func (q *queue) Poll() (interface{}, bool) {
	val, ok := <-q.dequeue
	return val, ok
}

func (q *queue) forceClose() {
	q.close <- true
	close(q.close)
}

func (q *queue) Close(timeOut time.Duration) {
	q.once.Do(func() {
		close(q.enqueue)
		if timeOut >= 0 {
			go func() {
				time.Sleep(timeOut)
				q.forceClose()
			}()
		}
	})
}

// New creates new instance of queue
func New() Queue {
	q := queue{
		enqueue: make(chan interface{}, 1),
		dequeue: make(chan interface{}, 1),
		close:   make(chan bool, 1),
	}
	go q.manage()

	return &q
}
