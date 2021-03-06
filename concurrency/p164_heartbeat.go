package concurrency

import (
	"fmt"
	"time"
)

func doWork(done <-chan interface{}, pulseInterval time.Duration) (<-chan interface{}, <-chan time.Time) {
	heartbeat := make(chan interface{}) // 1
	results := make(chan time.Time)

	go func() {
		defer close(heartbeat)
		defer close(results)

		pulse := time.Tick(pulseInterval)       // 2
		workGen := time.Tick(2 * pulseInterval) // 3

		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
			default: // 4
			}
		}
		sendResult := func(r time.Time) {
			for {
				select {
				case <-done:
					return
				case <-pulse: // 5
					sendPulse()
				case results <- r:
					return
				}
			}
		}

		for {
			select {
			case <-done:
				return
			case <-pulse: // 5
				sendPulse()
			case r := <-workGen:
				sendResult(r)
			}
		}
	}()

	return heartbeat, results
}

func doWorkBad(done <-chan interface{}, pulseInterval time.Duration) (<-chan interface{}, <-chan time.Time) {
	heartbeat := make(chan interface{}) // 1
	results := make(chan time.Time)

	go func() {
		defer close(heartbeat)
		defer close(results)

		pulse := time.Tick(pulseInterval)       // 2
		workGen := time.Tick(2 * pulseInterval) // 3

		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
			default: // 4
			}
		}
		sendResult := func(r time.Time) {
			for {
				select {
				case <-done:
					return
				case <-pulse: // 5
					sendPulse()
				case results <- r:
					return
				}
			}
		}

		for i := 0; i < 2; i++ {
			select {
			case <-done:
				return
			case <-pulse: // 5
				sendPulse()
			case r := <-workGen:
				sendResult(r)
			}
		}
	}()

	return heartbeat, results
}

func P164HeartBeat() {
	done := make(chan interface{})
	time.AfterFunc(10*time.Second, func() { close(done) })

	const timeout = 2 * time.Second
	heartbeat, results := doWorkBad(done, timeout/2)
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				fmt.Println("heartbeat die")
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if !ok {
				fmt.Println("results die")
				return
			}
			fmt.Printf("results %v\n", r.Second())
		case <-time.After(timeout):
			fmt.Println("worker goroutine is not healthy!")
			return
		}
	}
}
