package task_2

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func Start() {

	var (
		numOfChannels = rand.Intn(12) + 1
		channels      []chan interface{}
	)

	fmt.Println(fmt.Sprintf("creating %d channels..", numOfChannels))
	for i := 0; i < numOfChannels; i++ {
		channels = append(channels, make(chan interface{}))
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	joinedChannel := join(ctx, channels...)
	for i, c := range channels {
		c <- fmt.Sprint("hello_", i)
	}

	for {
		select {
		case <-time.After(10 * time.Second):
			fmt.Println("reached time limit waiting for joined output, exiting.")
			return
		case joinedOutput, ok := <-joinedChannel:
			if !ok {
				return
			}
			fmt.Println("joined output: ", joinedOutput)
		}
	}

}

func join(ctx context.Context, channels ...chan interface{}) chan interface{} {
	joinedChannel := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(len(channels))
	all := make([]interface{}, 0)
	for _, c := range channels {
		fmt.Println("dispatching goroutine..")
		go func(ctx context.Context, c <-chan interface{}) {
			val := <-c
			select {
			case <-time.After(time.Duration(rand.Intn(2)) * time.Second):
				all = append(all, val)
			case <-ctx.Done():
				wg.Done()
				return
			}

			wg.Done()
		}(ctx, c)
	}
	go func() {
		wg.Wait()
		joinedChannel <- all
		close(joinedChannel)
	}()
	return joinedChannel
}
