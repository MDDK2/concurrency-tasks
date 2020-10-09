package task_3

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

const maxJobs = 100

var (
	currentProcessingFiles = 0
	maxProcessingFiles     = 1
	jobQueue               = make(chan ProcessCSVRequest, maxJobs)
	stopFunc               context.CancelFunc
)

type CSVService struct {
}

type ProcessCSVRequest struct {
	ctx      context.Context
	filePath string
	handler  func(context.Context, []string) error
	result   chan error
}

func (s *CSVService) SetMaxProcessingFiles(val int) {
	maxProcessingFiles = val
}

func (s *CSVService) ProcessCSV(
	ctx context.Context,
	filePath string,
	handler func(context.Context, []string) error,
) chan error {
	result := make(chan error)

	jobQueue <- ProcessCSVRequest{
		ctx:      ctx,
		filePath: filePath,
		handler:  handler,
		result:   result,
	}

	fmt.Println("enqueued job")
	return result
}

func New() *CSVService {
	return &CSVService{}
}

func (s *CSVService) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	stopFunc = cancel
	go jobWorker(ctx, jobQueue)
}

func (s *CSVService) Stop() {
	stopFunc()
}

func jobWorker(ctx context.Context, jobChan <-chan ProcessCSVRequest) {
	for {
		select {
		case <-ctx.Done():
			return // for now just return
		case job := <-jobChan:
			fmt.Println("processing job :) ...")
			go processCSV(job.ctx, job.filePath, job.handler, job.result)
		default:
		}
	}
}

func processRecord(ctx context.Context, record []string, handler func(context.Context, []string) error) chan error {
	out := make(chan error)
	go func() {
		result := handler(ctx, record)
		out <- result
	}()
	return out
}
func processCSV(
	ctx context.Context,
	filePath string,
	handler func(context.Context, []string) error,
	result chan error,
) {

	if currentProcessingFiles == maxProcessingFiles {

	}
	// Open the file
	csvfile, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	// Parse the file
	r := csv.NewReader(csvfile)
	//r := csv.NewReader(bufio.NewReader(csvfile))

	// Iterate through the records
	//fmt.Println("<<<<<<<< INITIAL CSV CONTENTS >>>>>>>>")
	recordResultChans := make([]chan error, 0)
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("Record %s\n", record)

		recordResultChan := processRecord(ctx, record, handler)
		recordResultChans = append(recordResultChans, recordResultChan)
	}
	//fmt.Println(">>>>>>>>> INITIAL CSV CONTENTS END <<<<<<<<<<")

	for i := range recordResultChans {
		err := <-recordResultChans[i]
		if err != nil {
			result <- err
		}
	}
	result <- nil
}
