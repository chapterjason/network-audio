package circularbuffer

import (
	"sync"
)

// Queue is a thread-safe circular buffer.
type Queue struct {
	buffer  []any
	maxSize int

	// Ensures that Enqueue waits until the buffer is not full
	// and Dequeue waits until the buffer is not empty
	cond *sync.Cond

	readerMutex    sync.RWMutex
	readerPosition int

	writerMutex    sync.RWMutex
	writerPosition int

	fullMutex sync.RWMutex
	isFull    bool
}

// New creates a new Queue with the given maximum size.
func New(maxSize int) *Queue {
	if maxSize < 1 {
		panic("Invalid maxSize, should be at least 1")
	}

	queue := &Queue{maxSize: maxSize}
	queue.Clear()

	return queue
}

// Enqueue adds an item to the end of the queue. If the queue is full, the call blocks until an item is dequeued.
func (queue *Queue) Enqueue(value any) {
	queue.waitNotFull()

	queue.writerMutex.Lock()

	queue.buffer[queue.writerPosition] = value

	// Increment writerPosition
	queue.writerPosition = queue.writerPosition + 1

	// If the writerPosition is at the end of the buffer, set it to the beginning
	if queue.writerPosition >= queue.maxSize {
		queue.writerPosition = 0
	}

	queue.readerMutex.RLock()
	if queue.writerPosition == queue.readerPosition {
		queue.fullMutex.Lock()
		queue.isFull = true
		queue.fullMutex.Unlock()
	}
	queue.readerMutex.RUnlock()

	queue.writerMutex.Unlock()

	queue.cond.Broadcast()
}

// Dequeue removes an item from the beginning of the queue. If the queue is empty, the call blocks until an item is enqueued.
func (queue *Queue) Dequeue() any {
	queue.waitNotEmpty()

	queue.readerMutex.Lock()

	value := queue.buffer[queue.readerPosition]

	if value != nil {
		queue.buffer[queue.readerPosition] = nil
		queue.readerPosition = queue.readerPosition + 1

		if queue.readerPosition >= queue.maxSize {
			queue.readerPosition = 0
		}

		queue.fullMutex.Lock()
		queue.isFull = false
		queue.fullMutex.Unlock()
	}

	queue.readerMutex.Unlock()

	queue.cond.Broadcast()

	return value
}

// Peek returns the item at the beginning of the queue without removing it. If the queue is empty, the call blocks until an item is enqueued.
func (queue *Queue) Peek() any {
	queue.waitNotEmpty()

	queue.writerMutex.Lock()
	defer queue.writerMutex.Unlock()

	return queue.buffer[queue.readerPosition]
}

func (queue *Queue) Empty() bool {
	return queue.Size() == 0
}

func (queue *Queue) Full() bool {
	queue.fullMutex.RLock()
	defer queue.fullMutex.RUnlock()

	return queue.isFull
}

func (queue *Queue) Size() int {
	queue.readerMutex.RLock()
	defer queue.readerMutex.RUnlock()
	queue.writerMutex.RLock()
	defer queue.writerMutex.RUnlock()

	if queue.writerPosition < queue.readerPosition {
		return queue.maxSize - queue.readerPosition + queue.writerPosition
	} else if queue.writerPosition == queue.readerPosition {
		if queue.Full() {
			return queue.maxSize
		} else {
			return 0
		}
	}

	return queue.writerPosition - queue.readerPosition
}

func (queue *Queue) Clear() {
	queue.readerMutex.Lock()
	defer queue.readerMutex.Unlock()
	queue.writerMutex.Lock()
	defer queue.writerMutex.Unlock()
	queue.fullMutex.Lock()
	defer queue.fullMutex.Unlock()

	queue.buffer = make([]any, queue.maxSize, queue.maxSize)
	queue.cond = sync.NewCond(&sync.Mutex{})
	queue.readerPosition = 0
	queue.writerPosition = 0
	queue.isFull = false
}

func (queue *Queue) waitNotEmpty() {
	queue.cond.L.Lock()

	for queue.Empty() {
		queue.cond.Wait()
	}

	queue.cond.L.Unlock()
}

func (queue *Queue) waitNotFull() {
	queue.cond.L.Lock()

	for queue.Full() {
		queue.cond.Wait()
	}

	queue.cond.L.Unlock()
}
