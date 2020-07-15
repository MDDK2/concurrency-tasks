package task_1

import (
	"fmt"
	"math/rand"
	"time"
)

func Start() {
	type someStruct struct {
	}
	var (
		workDone        = make(chan string)
		stopChan        = make(chan struct{})
		startAgainChan  = make(chan struct{})
		numOfGoRoutines = rand.Intn(2) + 1
	)

	fmt.Println(fmt.Sprintf("creating %d goroutines..", numOfGoRoutines))
	for i := 0; i < numOfGoRoutines; i++ {
		go countElephants(i, workDone, stopChan, startAgainChan)
	}

	for i := 1; i <= 25; i++ {
		time.Sleep(1 * time.Second)
		fmt.Println("step: ", i)
		if i == 5 {
			fmt.Println("STOP!")
			close(stopChan)
		} else if i == 15 {
			fmt.Println("START AGAIN!")
			close(startAgainChan)
		} else {
			select {
			case workOutput := <-workDone:
				fmt.Println(workOutput)
			default:

			}

		}
	}

}

func countElephants(workerNumber int, output chan<- string, stopchan chan struct{}, stopStartChan chan struct{}) {
	var count int
	for {
		select {
		default:
			// doing some work
			time.Sleep(2 * time.Second)
			count++
			output <- fmt.Sprintf("worker: %d has counted: %d elephants..", workerNumber, count)
		case <-stopchan:
			// stop
			output <- fmt.Sprintf("worker: %d is stopping, waiting to be started again..", workerNumber)
			stopchan = nil
			<-stopStartChan
			output <- fmt.Sprintf("worker: %d is happily starting again..", workerNumber)
		}
	}
}
