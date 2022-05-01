package circularbuffer

import (
	"sync"
	"testing"
	"time"
)

func TestQueue_Clear(t *testing.T) {
	t.Run(
		"should clear the queue",
		func(t *testing.T) {
			q := New(10)

			if s := q.Size(); s != 0 {
				t.Fatal("queue is not empty")
			}

			q.Enqueue(1)

			if s := q.Size(); s != 1 {
				t.Fatal("queue size is not 1")
			}

			q.Clear()

			if s := q.Size(); s != 0 {
				t.Fatal("queue size is not 0")
			}
		},
	)
}

func TestQueue_Dequeue(t *testing.T) {
	t.Run(
		"should dequeue the first element",
		func(t *testing.T) {
			q := New(10)

			q.Enqueue(1)

			v := q.Dequeue()

			if v != 1 {
				t.Fatal("dequeue value is not 1")
			}

			if s := q.Size(); s != 0 {
				t.Fatal("queue size is not 0")
			}

			q.Enqueue(2)
			q.Enqueue(3)

			v = q.Dequeue()

			if v != 2 {
				t.Fatal("dequeue value is not 2")
			}

			if s := q.Size(); s != 1 {
				t.Fatal("queue size is not 1")
			}

			v = q.Dequeue()

			if v != 3 {
				t.Fatal("dequeue value is not 3")
			}

			if s := q.Size(); s != 0 {
				t.Fatal("queue size is not 0")
			}
		},
	)

	t.Run(
		"should wait for an element to be available",
		func(t *testing.T) {
			q := New(10)
			wg := &sync.WaitGroup{}
			var start, stop time.Time
			item := 0

			wg.Add(1)

			go func() {
				start = time.Now()
				item = q.Dequeue().(int)
				stop = time.Now()
				wg.Done()
			}()

			go func() {
				time.Sleep(time.Millisecond * 100)
				q.Enqueue(1)
			}()

			wg.Wait()

			if stop.Sub(start) < time.Millisecond*100 {
				t.Fatal("queue did not block")
			}

			if item != 1 {
				t.Fatal("dequeue value is not 1")
			}
		},
	)
}

func TestQueue_Empty(t *testing.T) {
	t.Run(
		"should return true if the queue is empty",
		func(t *testing.T) {
			q := New(5)

			// queue is initially empty
			if !q.Empty() {
				t.Fatal("queue is not empty")
			}

			q.Enqueue(1)

			// we added an element, so queue is not empty
			if q.Empty() {
				t.Fatal("queue is empty")
			}

			q.Dequeue()

			// we dequeued an element, so queue is empty again
			if !q.Empty() {
				t.Fatal("queue is not empty")
			}
		},
	)

	t.Run(
		"should return false if the queue is not empty",
		func(t *testing.T) {
			q := New(2)

			if !q.Empty() {
				t.Fatal("queue is not empty")
			}

			q.Enqueue(1)

			if q.Empty() {
				t.Fatal("queue is empty")
			}

			q.Enqueue(2)

			if q.Empty() {
				t.Fatal("queue is empty")
			}

			q.Dequeue()

			if q.Empty() {
				t.Fatal("queue is empty")
			}

			q.Dequeue()

			if !q.Empty() {
				t.Fatal("queue is not empty")
			}

		},
	)
}

func TestQueue_Enqueue(t *testing.T) {
	t.Run(
		"should enqueue an element",
		func(t *testing.T) {
			q := New(10)

			q.Enqueue(1)

			if s := q.Size(); s != 1 {
				t.Fatal("queue size is not 1")
			}

			q.Enqueue(2)

			if s := q.Size(); s != 2 {
				t.Fatal("queue size is not 2")
			}
		},
	)

	t.Run(
		"should block if the queue is full",
		func(t *testing.T) {
			q := New(1)
			var start, stop time.Time

			q.Enqueue(1)

			wg := &sync.WaitGroup{}
			wg.Add(1)

			go func() {
				start = time.Now()
				q.Enqueue(2)
				stop = time.Now()
				wg.Done()
			}()

			go func() {
				time.Sleep(time.Millisecond * 100)
				q.Dequeue()
			}()

			wg.Wait()

			if stop.Sub(start) < time.Millisecond*100 {
				t.Fatal("queue did not block")
			}

			if s := q.Size(); s != 1 {
				t.Fatal("queue size is not 1")
			}
		},
	)
}

func TestQueue_Full(t *testing.T) {
	t.Run(
		"should return true if the queue is full",
		func(t *testing.T) {
			q := New(1)

			if q.Full() {
				t.Fatal("queue is full")
			}

			q.Enqueue(1)

			if !q.Full() {
				t.Fatal("queue is not full")
			}

			q.Dequeue()

			if q.Full() {
				t.Fatal("queue is full")
			}
		},
	)
}

func TestQueue_Peek(t *testing.T) {
	t.Run(
		"should return the first element in the queue without removing it",
		func(t *testing.T) {
			q := New(10)

			q.Enqueue(1)

			if v := q.Peek(); v != 1 {
				t.Fatal("queue peek value is not 1")
			}

			q.Enqueue(2)

			if v := q.Peek(); v != 1 {
				t.Fatal("queue peek value is not 1")
			}

			q.Dequeue()

			if v := q.Peek(); v != 2 {
				t.Fatal("queue peek value is not 2")
			}
		},
	)
}

func TestQueue_Size(t *testing.T) {
	t.Run(
		"should return the size of the queue",
		func(t *testing.T) {
			q := New(10)

			if s := q.Size(); s != 0 {
				t.Fatal("queue size is not 0")
			}

			q.Enqueue(1)

			if s := q.Size(); s != 1 {
				t.Fatal("queue size is not 1")
			}

			q.Enqueue(2)

			if s := q.Size(); s != 2 {
				t.Fatal("queue size is not 2")
			}

			q.Dequeue()

			if s := q.Size(); s != 1 {
				t.Fatal("queue size is not 1")
			}

			q.Dequeue()

			if s := q.Size(); s != 0 {
				t.Fatal("queue size is not 0")
			}
		},
	)
}
