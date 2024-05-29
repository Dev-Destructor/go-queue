package mq

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"sync"
	"testing"
)

// goos: linux
// goarch: amd64
// pkg: go-queue/pkg/queue
// cpu: AMD Ryzen 5 3550H with Radeon Vega Mobile Gfx
// go test -cpu 1,2,4 -bench . -benchmem

func BenchmarkPublishToNSubscribers(b *testing.B, n int) {

	broker := NewBroker()
	defer broker.Close(0)

	reg := regexp.MustCompile(`test\.*`)

	for i := 0; i < n; i++ {
		go func() {
			subscriber := broker.Subscribe(reg)
			for _, ok := subscriber.Poll(); ok; _, ok = subscriber.Poll() {
			}
		}()
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := strconv.Itoa(100)
			broker.Publish("test"+id, rand.Intn(1000))
		}
	})
}

func BenchmarkSubscribeToNPublishers(b *testing.B, n int) {
	var wg sync.WaitGroup

	broker := NewBroker()
	defer broker.Close(0)

	stop := make(chan bool)

	subscriber := broker.Subscribe(regexp.MustCompile(`test\.*`))

	for i := 0; i < n; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					broker.Publish("test:"+strconv.Itoa(i), rand.Intn(1000))
				}
			}
		}()
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			subscriber.Poll()
		}
	})

	stop <- true
	close(stop)

	wg.Wait()
}

func BenchmarkMSubscriberNPublisher(b *testing.B, m, n int) {
	var wg sync.WaitGroup

	broker := NewBroker()
	defer broker.Close(0)

	stop := make(chan bool)
	subscriber := make([]Poller, m)
	for i := 0; i < m; i++ {
		subscriber[i] = broker.Subscribe(regexp.MustCompile(`test\.*`))
	}

	for i := 0; i < n; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					broker.Publish("test:"+strconv.Itoa(i), rand.Intn(1000))
				}
			}
		}()
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		sub := subscriber[rand.Intn(m)]
		for pb.Next() {
			sub.Poll()
		}
	})

	stop <- true
	close(stop)

	wg.Wait()
}

func BenchmarkPublish(b *testing.B) {
	// BenchmarkPublish/Subscribers=1           	 1386271	       823.4 ns/op	      92 B/op	       2 allocs/op
	// BenchmarkPublish/Subscribers=1-2         	 1000000	      1566 ns/op	     104 B/op	       2 allocs/op
	// BenchmarkPublish/Subscribers=1-4         	 1000000	      1211 ns/op	     105 B/op	       2 allocs/op
	// BenchmarkPublish/Subscribers=10          	 1000000	      5735 ns/op	     214 B/op	      11 allocs/op
	// BenchmarkPublish/Subscribers=10-2        	  227755	      4969 ns/op	     245 B/op	      12 allocs/op
	// BenchmarkPublish/Subscribers=10-4        	  367089	      3277 ns/op	     260 B/op	      11 allocs/op
	// BenchmarkPublish/Subscribers=30          	 1000000	     15261 ns/op	     545 B/op	      28 allocs/op
	// BenchmarkPublish/Subscribers=30-2        	   90852	     11443 ns/op	     646 B/op	      32 allocs/op
	// BenchmarkPublish/Subscribers=30-4        	  155415	      7421 ns/op	     678 B/op	      30 allocs/op
	// BenchmarkPublish/Subscribers=50          	 1000000	     26762 ns/op	     912 B/op	      46 allocs/op
	// BenchmarkPublish/Subscribers=50-2        	   59864	     18098 ns/op	    1046 B/op	      52 allocs/op
	// BenchmarkPublish/Subscribers=50-4        	   87877	     11622 ns/op	    1082 B/op	      50 allocs/op

	subscriberTopicRatios := []int{1, 10, 30, 50}

	for _, subscriberCount := range subscriberTopicRatios {
		name := fmt.Sprintf("Subscribers=%d", subscriberCount)
		b.Run(name, func(b *testing.B) {
			BenchmarkPublishToNSubscribers(b, subscriberCount)
		})
	}
}

func BenchmarkSubscribe(b *testing.B) {
	// BenchmarkSubscribe/Publishers=1           	 1413211	       885.3 ns/op	      97 B/op	       1 allocs/op
	// BenchmarkSubscribe/Publishers=1-2         	  817126	      1948 ns/op	      44 B/op	       1 allocs/op
	// BenchmarkSubscribe/Publishers=1-4         	  604620	      1983 ns/op	      37 B/op	       2 allocs/op
	// BenchmarkSubscribe/Publishers=10          	  839402	      1553 ns/op	     206 B/op	       3 allocs/op
	// BenchmarkSubscribe/Publishers=10-2        	  437983	      2951 ns/op	     612 B/op	      10 allocs/op
	// BenchmarkSubscribe/Publishers=10-4        	  620043	      2563 ns/op	     284 B/op	       5 allocs/op
	// BenchmarkSubscribe/Publishers=30          	  745581	      1790 ns/op	     224 B/op	       4 allocs/op
	// BenchmarkSubscribe/Publishers=30-2        	  245145	      7497 ns/op	    1740 B/op	      31 allocs/op
	// BenchmarkSubscribe/Publishers=30-4        	  475243	      3423 ns/op	     322 B/op	       5 allocs/op
	// BenchmarkSubscribe/Publishers=50          	  727684	      1546 ns/op	     211 B/op	       3 allocs/op
	// BenchmarkSubscribe/Publishers=50-2        	  166479	     11814 ns/op	    2771 B/op	      49 allocs/op
	// BenchmarkSubscribe/Publishers=50-4        	  339495	      3129 ns/op	     394 B/op	       6 allocs/op

	publisherCounts := []int{1, 10, 30, 50}

	for _, publisherCount := range publisherCounts {
		name := fmt.Sprintf("Publishers=%d", publisherCount)
		b.Run(name, func(b *testing.B) {
			BenchmarkSubscribeToNPublishers(b, publisherCount)
		})
	}
}

func BenchmarkMultiSubscriber(b *testing.B) {
	// BenchmarkMultiSubscriber/Subscribers=10,_Publishers=10           	  263068	      4993 ns/op	     817 B/op	       2 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=10,_Publishers=10-2         	  595820	      1700 ns/op	     417 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=10,_Publishers=10-4         	 1670811	       736.1 ns/op	     159 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=10,_Publishers=30           	  234508	      5252 ns/op	     913 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=10,_Publishers=30-2         	  552794	      1815 ns/op	     364 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=10,_Publishers=30-4         	 1398697	       863.2 ns/op	     258 B/op	       0 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=10,_Publishers=50           	  212618	      5395 ns/op	     806 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=10,_Publishers=50-2         	  743194	      3393 ns/op	     895 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=10,_Publishers=50-4         	 1778894	      1017 ns/op	     256 B/op	       0 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=30,_Publishers=10           	   76186	     14395 ns/op	    2730 B/op	       2 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=30,_Publishers=10-2         	  273032	      4215 ns/op	    1193 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=30,_Publishers=10-4         	  739215	      1648 ns/op	     660 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=30,_Publishers=30           	   77022	     14154 ns/op	    2703 B/op	       2 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=30,_Publishers=30-2         	  299362	      4289 ns/op	    1370 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=30,_Publishers=30-4         	  799321	      1964 ns/op	     796 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=30,_Publishers=50           	   79618	     14376 ns/op	    2618 B/op	       2 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=30,_Publishers=50-2         	  292317	      4406 ns/op	    1120 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=30,_Publishers=50-4         	  743510	      1548 ns/op	     661 B/op	       0 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=50,_Publishers=10           	   49104	     22600 ns/op	    4396 B/op	       2 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=50,_Publishers=10-2         	  181768	      6906 ns/op	    1897 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=50,_Publishers=10-4         	  521338	      2418 ns/op	    1025 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=50,_Publishers=30           	   50953	     24010 ns/op	    4241 B/op	       2 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=50,_Publishers=30-2         	  157732	      7897 ns/op	    2185 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=50,_Publishers=30-4         	  527623	      2368 ns/op	    1016 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=50,_Publishers=50           	   45751	     24467 ns/op	    3706 B/op	       2 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=50,_Publishers=50-2         	  164554	      7082 ns/op	    2097 B/op	       1 allocs/op
	// BenchmarkMultiSubscriber/Subscribers=50,_Publishers=50-4         	  544424	      2282 ns/op	     987 B/op	       0 allocs/op

	subscriberCounts := []int{10, 30, 50}
	publisherCount := []int{10, 30, 50}
	for _, subscriberCount := range subscriberCounts {
		for _, publisherCount := range publisherCount {
			name := fmt.Sprintf("Subscribers=%d, Publishers=%d", subscriberCount, publisherCount)
			b.Run(name, func(b *testing.B) {
				BenchmarkMSubscriberNPublisher(b, subscriberCount, publisherCount)
			})
		}
	}
}
