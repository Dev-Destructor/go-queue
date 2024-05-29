package queue

import (
	"math/rand"
	"testing"
)

// goos: linux
// goarch: amd64
// pkg: go-queue/pkg/queue
// cpu: AMD Ryzen 5 3550H with Radeon Vega Mobile Gfx
// go test -cpu 1,2,4 -bench . -benchmem

func BenchmarkPushWithoutPoll(b *testing.B) {

	queue := New()
	defer queue.Close(0)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			queue.Push(rand.Intn(1000000))
		}
	})
}

func BenchmarkBothPushPollWithPrefilledQueue(b *testing.B) {

	queue := New()

	defer queue.Close(0)

	for i := 0; i < 10000; i++ {
		queue.Push(rand.Intn(10000))
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := rand.Intn(10000)
			queue.Push(n)
			if n > 5000 && n < 10000 {
				queue.Poll()
			}
		}
	})
}

func BenchmarkPollWithAsyncPublish(b *testing.B) {

	queue := New()
	defer queue.Close(0)

	closeCh := make(chan bool)
	go func() {
		for {
			select {
			case <-closeCh:
				return
			default:
				queue.Push(rand.Intn(100000))
			}
		}
	}()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			queue.Poll()
		}
	})

	closeCh <- true
}
