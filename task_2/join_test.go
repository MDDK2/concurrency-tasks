package task_2

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	var (
		messagesOut1 = []int{11, 12, 13, 14, 15, 16, 17}
		messagesOut2 = []int{21, 22, 23, 24, 25, 26, 27}
		messagesOut3 = []int{31, 32, 33, 34, 35, 36, 37}
		messagesOut4 = []int{41, 42, 43, 44, 45, 46, 47}
		messagesOut5 = []int{51, 52, 53, 54, 55, 56, 57}
		messagesOut6 = []int{61, 62, 63, 64, 65, 66, 67}
		messagesOut  = [][]int{
			messagesOut1,
			messagesOut2,
			messagesOut3,
			messagesOut4,
			messagesOut5,
			messagesOut6,
		}

		expectedJoinedOutput1 = []interface{}{11, 21, 31, 41, 51, 61}
		expectedJoinedOutput2 = []interface{}{12, 22, 32, 42, 52, 62}
		expectedJoinedOutput3 = []interface{}{13, 23, 33, 43, 53, 63}
		expectedJoinedOutput4 = []interface{}{14, 24, 34, 44, 54, 64}
		expectedJoinedOutput5 = []interface{}{15, 25, 35, 45, 55, 65}
		expectedJoinedOutput6 = []interface{}{16, 26, 36, 46, 56, 66}
		expectedJoinedOutput7 = []interface{}{17, 27, 37, 47, 57, 67}
		expectedJoinedOutputs = [][]interface{}{
			expectedJoinedOutput1,
			expectedJoinedOutput2,
			expectedJoinedOutput3,
			expectedJoinedOutput4,
			expectedJoinedOutput5,
			expectedJoinedOutput6,
			expectedJoinedOutput7,
		}

		numOfChannels = 6
		channels      []chan interface{}
	)

	fmt.Println(fmt.Sprintf("TEST: creating %d channels..", numOfChannels))
	for i := 0; i < numOfChannels; i++ {
		channels = append(channels, make(chan interface{}, 10))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)

	joinedChan := join(ctx, channels...)

	for i := range channels {
		var maxRandomDelayMS = 1000
		if i == 4 {
			maxRandomDelayMS = 5000 // make one a lot slower to see if others wait correctly
		}
		go writeToChannel(ctx, channels[i], messagesOut[i], maxRandomDelayMS, i)
	}

	var joinedOutputsReceived = 0

	for {
		select {
		case <-ctx.Done():
			return
		case joinedOutput, ok := <-joinedChan:
			if !ok {
				return
			}
			joinedOutputsReceived++
			fmt.Println(fmt.Printf("TEST: RECEIVED JOINED OUTPUT (iteration: %d): %v", joinedOutputsReceived, joinedOutput))
			assert.ElementsMatch(t, expectedJoinedOutputs[joinedOutputsReceived-1], joinedOutput)
		}
		if joinedOutputsReceived == len(expectedJoinedOutputs) {
			cancel()
		}
	}
}

func writeToChannel(ctx context.Context, out chan interface{}, messages []int, randomDelayMS int, debugChannelIndex int) {
	for _, m := range messages {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(rand.Intn(randomDelayMS)) * time.Millisecond):
			select {
			case <-ctx.Done():
				return
			case out <- m:
				fmt.Println(fmt.Sprintf("TEST: sent out m: %v to channel %d", m, debugChannelIndex))
			}
		}
	}
}
