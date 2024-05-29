package queue

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	queue := New()
	for i := 1; i <= 5; i++ {
		queue.Push(i)
	}

	queue.Close(-1)

	expected := []int{1, 2, 3, 4, 5}
	for i := 0; i < 5; i++ {
		val, ok := queue.Poll()
		if !ok {
			t.Fatalf("No more values to poll, but expected %d\n", expected[i])
		}
		if val != expected[i] {
			t.Errorf("Invalid Value: Expected: %d, Obtained: %v\n", expected[i], val)
		}
	}

	_, ok := queue.Poll()
	if ok {
		t.Errorf("Poll on available value should be False Got True\n")
	}
}

func TestPollAsync(t *testing.T) {
	queue := New()
	wg := sync.WaitGroup{}

	wg.Add(1)
	lastValue := 0

	go func() {
		defer wg.Done()

		for value, ok := queue.Poll(); ok; value, ok = queue.Poll() {
			if value != lastValue {
				t.Errorf("Invalid Value Obtained: Last: %v, Current: %v\n", lastValue, value)
			}
			lastValue++
		}
	}()

	maxValue := 100
	for i := 0; i < maxValue; i++ {
		queue.Push(i)
	}

	queue.Close(-1)

	wg.Wait()

	if lastValue != 100 {
		t.Errorf("Invalid Last Value Obtained: %v\n", lastValue)
	}
}

func TestPushAsync(t *testing.T) {
	queue := New()

	go func() {
		maxValue := 100
		for i := 0; i < maxValue; i++ {
			queue.Push(i)
		}
		queue.Close(-1)
	}()

	lastValue := 0
	for value, ok := queue.Poll(); ok; value, ok = queue.Poll() {
		if value != lastValue {
			t.Errorf("Invalid Value Obtained: Last: %v, Current: %v\n", lastValue, value)
		}
		lastValue++
	}

	if lastValue != 100 {
		t.Errorf("Invalid Last Value Obtained: %v\n", lastValue)
	}
}

func TestPushPollSequential(t *testing.T) {
	queue := New()

	maxValue := 100
	for i := 0; i < maxValue; i++ {
		queue.Push(i)
	}
	queue.Close(-1)

	lastValue := 0
	for value, ok := queue.Poll(); ok; value, ok = queue.Poll() {
		if value != lastValue {
			t.Errorf("Invalid Value Obtained: Last: %v, Current: %v\n", lastValue, value)
		}
		lastValue++
	}

	if lastValue != 100 {
		t.Errorf("Invalid Last Value Obtained: %v\n", lastValue)
	}
}

func TestRoutineClose(t *testing.T) {
	expected := runtime.NumGoroutine()
	queue := New()

	for i := 0; i < 10; i++ {
		queue.Push(i)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		queue.Poll()
	}()

	queue.Close(0)

	wg.Wait()
	time.Sleep(time.Millisecond)

	if current := runtime.NumGoroutine(); current != expected {
		t.Errorf("Invalid Number of Goroutines: Expected: %d, Obtained: %d\n", expected, current)
	}
}
