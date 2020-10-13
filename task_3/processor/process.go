package processor

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/concurrency-tasks/task_3/errors"
	"github.com/concurrency-tasks/task_3/persist"
)

type Processor struct {
	wg                     *sync.WaitGroup
	currentProcessingFiles int
}

type ProcessCSVRequest struct {
	Ctx                     context.Context
	ID                      string
	FilePath                string
	OutputFilePath          string
	Handler                 func(context.Context, []string) ProcessRecordResponse
	ResultChan              chan error
	AlreadyProcessedRecords [][]string
}

type ProcessCSVRequestSave struct {
	ID               string
	FilePath         string
	OutputFilePath   string
	ProcessedRecords [][]string
}

type ProcessRecordResponse struct {
	Record []string
	Err    error
}

func (p *Processor) GetCurrentProcessingFiles() int {
	return p.currentProcessingFiles
}

func New(wg *sync.WaitGroup) *Processor {
	return &Processor{
		wg: wg,
	}
}

func processRecord(
	ctx context.Context,
	record []string,
	handler func(context.Context, []string) ProcessRecordResponse) chan ProcessRecordResponse {
	out := make(chan ProcessRecordResponse)
	go func() {
		result := handler(ctx, record)
		out <- result
	}()
	return out
}

func (p *Processor) ProcessCSV(
	ctx context.Context,
	id string,
	filePath string,
	outputFilePath string,
	alreadyProcessedRecords [][]string,
	handler func(context.Context, []string) ProcessRecordResponse,
	resultChan chan error,
	pauseChan chan bool,
) {
	fmt.Println("\n------------------------------------------")
	p.wg.Add(1)
	defer p.wg.Done()
	fmt.Println("processCSV(): opening csv file: ", filePath)
	csvfile, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("processCSV(): couldn't open the csv file ", err)
	}
	p.currentProcessingFiles++
	r := csv.NewReader(csvfile)

	recordResultChans := make([]chan ProcessRecordResponse, 0)
	var recordIndex int
	var numOfResultsToReceive int
	var debugProcessed []int
	var debugUnprocessed []int
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		alreadyProcessed := len(alreadyProcessedRecords) >= (recordIndex+1) && alreadyProcessedRecords[recordIndex] != nil
		if alreadyProcessed {
			recordResultChans = append(recordResultChans, nil)
			debugProcessed = append(debugProcessed, recordIndex)

		} else {
			recordResultChan := processRecord(ctx, record, handler)
			recordResultChans = append(recordResultChans, recordResultChan)
			numOfResultsToReceive++
			debugUnprocessed = append(debugUnprocessed, recordIndex)
		}
		recordIndex++

	}
	fmt.Println("processCSV(): records already processed: ", debugProcessed)
	fmt.Println("processCSV(): records NOT already processed: ", debugUnprocessed)

	processedRecords := make([][]string, len(recordResultChans))
	//fmt.Printf("processCSV(): already processed %d/%d:  \n", len(alreadyProcessedRecords), numOfResultsToReceive)
	fmt.Println("------------------------------------------\n ")
	copy(processedRecords, alreadyProcessedRecords)
	var (
		numOfResultsFromChannels int
		processedAll             bool
	)
	//fmt.Println("not processed, queued L: ", numOfResultsToReceive)
records:
	for {
		for i := range recordResultChans {
			select {
			case <-ctx.Done():
				break records
			case <-pauseChan:
				break records
			case res := <-recordResultChans[i]:
				fmt.Println("processCSV(): +1 processed record")
				if res.Err != nil {
					resultChan <- err
				}
				processedRecords[i] = res.Record
				numOfResultsFromChannels++
				if numOfResultsFromChannels == numOfResultsToReceive {
					processedAll = true
					break records
				}
			default:
			}
		}
	}

	// processedRecordsChan := join(ctx, recordResultChans)
	// responses := <-processedRecordsChan
	// for i := range responses {
	// 	res := res[i].(ProcessRecordResponse)
	// 	if res.err != nil {
	// 		// do something
	// 		fmt.Println("got err")
	// 	}
	// 	processedRecords = append(processedRecords, res.record)
	// }
	if processedAll {
		fmt.Println("processCSV(): Yey, processed all records!")
		saveCSV(processedRecords, outputFilePath)
		resultChan <- nil
	} else {
		fmt.Printf("processCSV(): aborted, processed %d/%d, saving progess: %s.tmp \n", numOfResultsFromChannels, numOfResultsToReceive, id)
		p := &ProcessCSVRequestSave{
			ID:               id,
			FilePath:         fmt.Sprintf("./storage/wip/%s", filePath),
			OutputFilePath:   outputFilePath,
			ProcessedRecords: processedRecords,
		}
		if err := persist.Save(fmt.Sprintf("./storage/wip/%s.tmp", id), p); err != nil {
			log.Fatalln(err)
		}
		resultChan <- errors.Interrupted
		saveCSV(processedRecords, fmt.Sprintf("./storage/wip/%s.csv", id))
	}
	p.currentProcessingFiles--

	// if err := persist.Save("./file2.tmp", p); err != nil {
	// 	log.Fatal(err)
	// }

	//

	// var network bytes.Buffer        // Stand-in for a network connection
	// enc := gob.NewEncoder(&network) // Will write to network.
	// //dec := gob.NewDecoder(&network) // Will read from network.

	// // Encode (send) some values.
	// err = enc.Encode(p)
	// if err != nil {
	// 	log.Fatal("encode error:", err)
	// }

	//

}

func saveCSV(data [][]string, outputFilePath string) {

	file, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatal(err) // return to client later instead internal error
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write(value)
		if err != nil {
			log.Fatal(err) // return to client later instead internal error
		}
	}
}
