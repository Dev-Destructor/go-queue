# Go-Queue

[![Test Cases](https://github.com/Dev-Destructor/go-queue/actions/workflows/ci.yaml/badge.svg)](https://github.com/Dev-Destructor/go-queue/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dev-Destructor/go-queue)](https://goreportcard.com/report/github.com/Dev-Destructor/go-queue)
[![Coverage Status](https://coveralls.io/repos/github.com/Dev-Destructor/go-queue/badge.svg?branch=master)](https://coveralls.io/github.com/Dev-Destructor/go-queue?branch=master)

An in-memory message broker build with Go.

### Description

Go-Queue is an in-memory message broker which is thread safe and is Interface based. It provides a comprehensive way of subscription to topics based on pattern match.

### Installation

```shell
go get github.com/Dev-Destructor/go-queue
```

### Sample Demo

```go
  func main() {
    broker := mq.NewBroker()
  }
```

#### Subscription of a topic

##### Through Exact matching

```go
  func main() {
    broker := mq.NewBroker()

    testSubscriber := broker.Subscribe(mq.ExactMatcher("test"))
  }
```

##### Through Exact matching

```go
  func main() {
    broker := mq.NewBroker()

    testSubscribers := broker.Subscribe(regexp.MustCompile(`tests\.\w*`))
  }
```

#### Publishing to a topic

```go
  func main() {
    broker := mq.NewBroker()

    broker.Publish("test", "Hello World")
  }
```

### Reading from a topic

```go
  func main() {
    broker := mq.NewBroker()

    testSubscriber := broker.Subscribe(mq.ExactMatcher("test"))

    go func() {
      value, ok := testSubscriber.Poll()
      if !ok {
        return
      }

      fmt.Println(value)
    }()
  }
```
