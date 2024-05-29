package mq

import (
	"fmt"
	"sync"
	"testing"
)

func TestBrokerOnSingleRoutine(t *testing.T) {
	broker := NewBroker()
	defer broker.Close(0)

	subscriber := broker.Subscribe(ExactMatcher("test"))

	maxCount := 100

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for expected := 0; expected < maxCount; expected++ {
			val, ok := subscriber.Poll()
			if !ok {
				t.Errorf("Poll on available value should be True got False")
			}
			if expected != val.(int) {
				t.Errorf("Invalid Value: Expected: %d Obtained: %v", expected, val)
			}
		}
	}()

	for i := 0; i < maxCount; i++ {
		broker.Publish("test", i)
	}

	wg.Wait()
}

func TestBrokerOnMultiRoutine(t *testing.T) {
	broker := NewBroker()
	defer broker.Close(0)

	subscriber := broker.Subscribe(ExactMatcher("test"))

	maxCount := 100

	routineCount := 10

	wg := sync.WaitGroup{}

	for i := 0; i < routineCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			last := int32(-1)
			for i := 0; i < maxCount/routineCount; i++ {
				val, ok := subscriber.Poll()

				if !ok {
					t.Errorf("Poll on available value should be True Got False")
				}
				if last > val.(int32) {
					t.Errorf("Invalid Value: Last Value: %d Obtained: %v", last, val)
				}

				last = val.(int32)
			}
		}()
	}

	for i := 0; i < maxCount; i++ {
		broker.Publish("test", int32(i))
	}

	wg.Wait()
}

func TestBrokerPollAfterClose(t *testing.T) {
	broker := NewBroker()
	subscriber := broker.Subscribe(ExactMatcher("test"))
	broker.Publish("test", "test value")

	{
		val, ok := subscriber.Poll()
		if val != "test value" {
			t.Errorf("Expected Value: test-value, Obtained: %v\n", val)
		}
		if !ok {
			t.Error("Poll should be True")
		}
	}

	broker.Close(-1)

	{
		_, ok := subscriber.Poll()
		if ok {
			t.Error("Poll should be False")
		}
	}

	{
		_, ok := subscriber.Poll()
		if ok {
			t.Error("Poll should be False")
		}
	}
}

func TestCloseBroker(t *testing.T) {

	broker := NewBroker()

	topics := []string{"topic1", "topic2", "topic3"}

	wg := sync.WaitGroup{}
	for _, topic := range topics {
		topic := topic
		subscriber := broker.Subscribe(ExactMatcher(topic))
		wg.Add(1)
		go func() {
			defer wg.Done()
			incr := 0
			for val, ok := subscriber.Poll(); ok; val, ok = subscriber.Poll() {
				expected := fmt.Sprintf("%s:%v", topic, incr)
				if expected != val.(string) {
					t.Errorf("Invalid Value: Expected: %v Obtained: %v", expected, val)
				}
				incr++
			}
		}()
	}

	maxCount := 100
	for _, topic := range topics {
		for i := 0; i < maxCount; i++ {
			broker.Publish(topic, fmt.Sprintf("%s:%v", topic, i))
		}
	}

	broker.Close(-1)
	wg.Wait()
}
