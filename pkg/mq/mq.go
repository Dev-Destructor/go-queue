package mq

import (
	"sync"
	"time"

	"github.com/Dev-Destructor/go-queue/pkg/queue"
)

// ExactMatcher accepts string as a parameter
type ExactMatcher string

// Matcher is an interface for matchString
type Matcher interface {

	// MatchString returns true if the pattern matches the string
	MatchString(string) bool
}

type queueMatcher struct {
	queue   queue.Queue
	matcher Matcher
}

type broker struct {
	queueMatchers []queueMatcher

	// ~11.5% faster operation speed while caching the matchers
	matchCache map[string]map[Matcher]bool
	sync.RWMutex
}

// Poller is wrapper for Poll function
type Poller interface {

	// Poller reads the data from queue and returns the value.
	// It will wait till there is consumable data.
	// If the resource is closed, then Poller will return,
	Poll() (interface{}, bool)
}

// Broker is the broker for interaction
type Broker interface {

	// Publish publishes data to a specific topic.
	Publish(topic string, data interface{})

	// Subscribe creates a Poller which polls data from matched topics.
	Subscribe(topic Matcher) Poller

	// CloseTopic closes the topic and removes the topic from the broker.
	// If the timeOut is less than 0, then all the resources will be read-only.
	CloseTopic(topic Matcher, timeOut time.Duration)

	// Close closes the broker and changes it to read only.
	// If the timeOut is less than 0, then all the resources will be read-only.
	Close(timeOut time.Duration)
}

// MatchString returns true if the pattern matches the string
func (em ExactMatcher) MatchString(pattern string) bool {
	return string(em) == pattern
}

func (b *broker) Publish(topic string, data interface{}) {
	b.RLock()
	matchers, ok := b.matchCache[topic]
	b.RUnlock()

	if !ok {
		b.Lock()
		matchers = make(map[Matcher]bool)
		for _, q := range b.queueMatchers {
			matchers[q.matcher] = q.matcher.MatchString(topic)
		}
		b.matchCache[topic] = matchers
		b.Unlock()
	}

	for _, q := range b.queueMatchers {
		if matchers[q.matcher] {
			q.queue.Push(data)
		}
	}
}

func (b *broker) Subscribe(matcher Matcher) Poller {
	b.Lock()
	defer b.Unlock()

	q := queue.New()
	b.queueMatchers = append(b.queueMatchers, queueMatcher{queue: q, matcher: matcher})

	b.matchCache = make(map[string]map[Matcher]bool)

	return q
}

func (b *broker) CloseTopic(matcher Matcher, timeOut time.Duration) {
	b.Lock()
	defer b.Unlock()

	for i, qm := range b.queueMatchers {
		if qm.matcher == matcher {
			qm.queue.Close(timeOut)
			b.queueMatchers = append(b.queueMatchers[:i], b.queueMatchers[i+1:]...)
			break
		}
	}
}

func (b *broker) Close(timeOut time.Duration) {
	b.Lock()
	defer b.Unlock()
	for _, qm := range b.queueMatchers {
		qm.queue.Close(timeOut)
	}

	b.queueMatchers = []queueMatcher{}
}

// NewBroker creates an instance of broker
func NewBroker() Broker {
	return &broker{
		queueMatchers: []queueMatcher{},
		matchCache:    make(map[string]map[Matcher]bool),
	}
}
