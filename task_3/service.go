package task_3

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	customerrors "github.com/concurrency-tasks/task_3/errors"
	"github.com/concurrency-tasks/task_3/persist"
	"github.com/concurrency-tasks/task_3/processor"
)

const maxJobsInQueue = 100

var (
	wg        sync.WaitGroup
	jobChan   chan processor.ProcessCSVRequest
	pauseChan chan bool
)

type Status string

const (
	starting Status = "starting"
	running  Status = "running"
	stopping Status = "stopping"
	stopped  Status = "stopped"
)

type CSVService struct {
	Status             Status
	processor          *processor.Processor
	maxProcessingFiles int
}

func New() *CSVService {
	processor := processor.New(&wg)
	return &CSVService{
		Status:             stopped,
		processor:          processor,
		maxProcessingFiles: 100,
	}
}

func (s *CSVService) SetMaxProcessingFiles(val int) {
	s.maxProcessingFiles = val
}

func (s *CSVService) GetMaxProcessingFiles() int {
	return s.maxProcessingFiles
}

func (s *CSVService) GetCurrentProcessingFiles() int {
	return s.processor.GetCurrentProcessingFiles()
}

func (s *CSVService) ProcessCSV(
	ctx context.Context,
	id string,
	filePath string,
	outputFilePath string,
	handler func(context.Context, []string) processor.ProcessRecordResponse,
) chan error {
	var (
		resultChan       = make(chan error, 1)
		output           = outputFilePath
		processedRecords [][]string
	)
	fmt.Println("\n------------------------------------------")
	fmt.Println("ProcessCSV(): received process csv requests, server status: ", s.Status)
	switch s.Status {
	case stopping:
		resultChan <- customerrors.Shutdown
		return resultChan
	case stopped:
		resultChan <- customerrors.Stopped
		return resultChan
	}
	wip := tryLoadWorkInProgress(id)
	if wip != nil {
		output = wip.FilePath
		processedRecords = wip.ProcessedRecords
	}

	job := processor.ProcessCSVRequest{
		Ctx:                     ctx,
		ID:                      id,
		FilePath:                filePath,
		OutputFilePath:          output,
		Handler:                 handler,
		ResultChan:              resultChan,
		AlreadyProcessedRecords: processedRecords,
	}
	//fmt.Println("ProcessCSV(): JOB: ", job)
	if err := tryEnqueue(job); err != nil {
		resultChan <- err
	}
	fmt.Println("------------------------------------------\n ")
	return resultChan
}

func tryLoadWorkInProgress(id string) *processor.ProcessCSVRequestSave {
	p := &processor.ProcessCSVRequestSave{}
	err := persist.Load(fmt.Sprintf("./storage/wip/%s.tmp", id), p)
	if os.IsNotExist(err) {
		fmt.Println("ProcessCSV() -> tryLoadWorkInProgress(): NOT FOUND [X]")
		return nil
	}
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("ProcessCSV() -> tryLoadWorkInProgress(): FOUND [+1]")
	return p
}

func (s *CSVService) Start() error {
	if s.Status != stopped {
		return errors.New(fmt.Sprint("service needs to be stopped in order to be started, status: ", s.Status))
	}
	s.Status = starting // probably not actually concurrent safe?
	jobChan = make(chan processor.ProcessCSVRequest, maxJobsInQueue)
	pauseChan = make(chan bool)
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	// go func() {
	// 	pause := <-pauseChan
	// 	if pause {
	// 		fmt.Println("pause exec")
	// 		s.stopService()
	// 		//stopFunc()
	// 		// replace this with exit once stopSercice() returns chan response
	// 		// about having finished cleanup
	// 		// time.Sleep(2 * time.Second)
	// 		// os.Exit(0)
	// 	}
	// }()
	go func() {
		sig := <-signalChan
		fmt.Println("received signal: ", sig)
		if s.Status == running {
			s.stopService()
		} else {
			fmt.Println("no need to stop service, status is already: ", s.Status)
		}
	}()
	fmt.Println("start job dispatcher")
	go s.jobDispatcher(jobChan)
	go debugNumOfGoroutines()
	s.Status = running
	return nil
}

// func (s *CSVService) Stop() {
// 	stopService()
// }

func debugNumOfGoroutines() {
	for {
		time.Sleep(2 * time.Second)
		fmt.Println("num of goroutines: ", runtime.NumGoroutine())
	}
}
func (s *CSVService) stopService() {
	if s.Status == running {
		s.Status = stopping
		fmt.Println("close pauseChan")
		close(pauseChan)
		close(jobChan)

		wg.Wait()

		fmt.Println("Finished cleanup")
		s.Status = stopped
		//os.Exit(0)
	}
	//when we finished cleanup, close jobChannel
}

func tryEnqueue(job processor.ProcessCSVRequest) error {
	select {
	case <-pauseChan:
		fmt.Println("ProcessCSV() -> tryEnqueue(): ", customerrors.Shutdown)
		return customerrors.Shutdown
	default:
		jobChan <- job
		fmt.Println("ProcessCSV() -> tryEnqueue(): ENQUEUED[+1]")
	}
	return nil
}

func (s *CSVService) jobDispatcher(jobChan <-chan processor.ProcessCSVRequest) {
	wg.Add(1)
	defer wg.Done()
	for {
		if s.processor.GetCurrentProcessingFiles() >= s.GetMaxProcessingFiles() {
			continue
		}
		select {
		case <-pauseChan: // process all remaining jobs and return
			fmt.Printf("jobDispatcher(): pauseChan CLOSED! process remaining jobs: %d, and return \n", len(jobChan))
			for job := range jobChan {
				s.processJob(job)
			}
			return
		case job, ok := <-jobChan:
			if ok {
				s.processJob(job)
			}
		default:
		}
	}
}

func (s *CSVService) processJob(job processor.ProcessCSVRequest) {
	fmt.Println("jobDispatcher(): -> processCSV()")
	go s.processor.ProcessCSV(
		job.Ctx,
		job.ID,
		job.FilePath,
		job.OutputFilePath,
		job.AlreadyProcessedRecords,
		job.Handler,
		job.ResultChan,
		pauseChan,
	)
}
