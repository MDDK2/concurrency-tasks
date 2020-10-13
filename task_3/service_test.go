package task_3

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/concurrency-tasks/task_3/processor"
	"github.com/stretchr/testify/assert"
)

var alterWithLongDelay = func(ctx context.Context, record []string) processor.ProcessRecordResponse {
	rand.Seed(time.Now().UnixNano())
	n := 1 + rand.Intn(7)
	//fmt.Printf("Sleeping %d seconds...\n", n)
	//time.Sleep(time.Duration(n) * time.Second)
	select {
	case <-ctx.Done():
		return processor.ProcessRecordResponse{} // could use 2 args here instead of object, wrap into object later in process code, then just retun nil, nil here
	case <-time.After(time.Duration(n) * time.Second):
		processedRecord := make([]string, len(record))
		for i := range record {
			processedRecord[i] = "altered intensively!: " + record[i]
		}
		return processor.ProcessRecordResponse{
			Record: processedRecord,
			Err:    nil,
		}
	}
}

var alterWithShortDelay = func(ctx context.Context, record []string) processor.ProcessRecordResponse {
	//rand.Seed(time.Now().UnixNano())
	//time.Sleep(time.Duration(1) * time.Second)
	select {
	case <-ctx.Done():
		return processor.ProcessRecordResponse{}
	case <-time.After(time.Duration(1) * time.Second):
		processedRecord := make([]string, len(record))
		for i := range record {
			processedRecord[i] = "altered lightly: " + record[i]
		}
		return processor.ProcessRecordResponse{
			Record: processedRecord,
			Err:    nil,
		}
	}
}

// func TestKillAndResume(t *testing.T) {
// 	service := New()
// 	service.Start()

// 	id := uuid.Must(uuid.NewV4()).String()
// 	resultChan1 := service.ProcessCSV(
// 		context.Background(),
// 		id,
// 		"test_game_data.csv",
// 		"add_words_intensive_work_to_every_field.csv",
// 		alterWithLongDelay,
// 	)

// 	go func() {
// 		time.Sleep(5 * time.Second)
// 		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
// 	}()

// 	wrapTestLog("ProcessCSV() called, waiting for 1st response.. ")
// 	errResponse1 := <-resultChan1
// 	wrapTestLog("response1 (err): ", errResponse1)

// 	service.Start()

// 	for tries := 0; tries < 10; tries++ {
// 		resultChan2 := service.ProcessCSV(
// 			context.Background(),
// 			id,
// 			"test_game_data.csv",
// 			"add_words_intensive_work_to_every_field.csv",
// 			alterWithLongDelay,
// 		)
// 		wrapTestLog("ProcessCSV() called again, waiting for 2nd response..")

// 		errResponse2 := <-resultChan2
// 		wrapTestLog("response2 (err): ", errResponse2)
// 		if errResponse2 == nil {
// 			break
// 		}
// 		time.Sleep(1 * time.Second)
// 	}
// }

// 	// fmt.Println("---------")
// 	// assert.True(t, true)
// 	// resultChan2 := service.ProcessCSV(
// 	// 	context.Background(),
// 	// "00799745-1f80-4b13-bd7f-f34304159d49",
// 	// 	"test_game_data.csv",
// 	// 	"test_game_data_work.csv",
// 	// 	func(ctx context.Context, record []string) ProcessRecordResponse {
// 	// 		// work
// 	// 		time.Sleep(time.Millisecond * 5)
// 	// 		processedRecord := make([]string, len(record))
// 	// 		for i := range record {
// 	// 			processedRecord[i] = "altered: " + record[i]
// 	// 		}
// 	// 		return ProcessRecordResponse{
// 	// 			record: processedRecord,
// 	// 			err:    nil,
// 	// 		}
// 	// 		// return errors.New()
// 	// 	})

// 	//time.Sleep(time.Second * 2)
// 	// GetCSVRecords("test_game_data.csv")

// 	// results := make([]error, 0, 2)
// 	// for {
// 	// 	if len(results) == 1 { // 2
// 	// 		break
// 	// 	}

// 	// 	select {
// 	// 	case r1 := <-resultChan1:
// 	// 		fmt.Println("r1: ", r1)
// 	// 		results = append(results, r1)
// 	// 	// case r2 := <-resultChan2:
// 	// 	// 	fmt.Println("r2: ", r2)
// 	// 	// 	results = append(results, r2)
// 	// 	default:
// 	// 	}

// 	// }

// }

// func TestKillAndResume_MultipleCSVs(t *testing.T) {
// 	service := New()
// 	service.Start()

// 	requestIDs := []string{
// 		"be1b5c41-f7fd-4732-8393-409226c7ce67",
// 		"06961b7c-6b7f-4bfb-a37e-2ef15fac6c81",
// 		"e8a84bc2-39f5-4dd7-aee4-78ab71c0d144",
// 		"3a288a96-28d4-42e4-b990-0a2a637bac1b",
// 		"5935bb34-9ecd-41c2-ae9c-ce70dc400cc8",
// 		"8ac99dd1-4b47-4116-9651-8fd22fe3152a",
// 		"2c9e3e92-3a61-4e3a-a5c6-ab9dbead421b",
// 		"16e9c27b-3173-4f3b-b7fb-08a64b058af4",
// 		"eca2ec08-09a2-4d26-b9cc-c5c2812861e9",
// 		"269ee3fe-5e30-42d8-a6a8-2f2e59f8d5d4",
// 	}

// 	resultChans := testProcessMultipleCSVs(service, requestIDs)

// 	go func() {
// 		time.Sleep(5 * time.Second)
// 		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
// 	}()

// 	for i, r := range resultChans {
// 		resp := <-r
// 		wrapTestLog("response from chan: ", i, "resp: ", resp)
// 	}
// 	// wrapTestLog("ProcessCSV() called, waiting for 1st response.. ")
// 	// errResponse1 := <-resultChan1
// 	// wrapTestLog("response1 (err): ", errResponse1)

// 	// errResponse1B := <-resultChan1B
// 	// wrapTestLog("response1B (err): ", errResponse1B)

// 	for tries := 0; tries < 10; tries++ {
// 		time.Sleep(1 * time.Second)
// 		err := service.Start()
// 		if err != nil {
// 			wrapTestLog("service.Start(): ", err)
// 			continue
// 		}
// 		resultChan2 := service.ProcessCSV(
// 			context.Background(),
// 			requestIDs[0],
// 			"test_game_data.csv",
// 			"add_words_intensive_work_to_every_field.csv",
// 			alterWithLongDelay,
// 		)
// 		wrapTestLog("ProcessCSV() called again, waiting for 2nd response..")

// 		errResponse2 := <-resultChan2
// 		wrapTestLog("response2 from chan: ", 0, "resp: ", errResponse2)
// 		if errResponse2 == nil {
// 			break
// 		}

// 	}

// 	resultChans2 := testProcessMultipleCSVs(service, requestIDs)
// 	for i, r := range resultChans2 {
// 		if i == 0 {
// 			continue
// 		}
// 		resp := <-r
// 		assert.NoError(t, resp)
// 		wrapTestLog("response from chan: ", i, "resp: ", resp)
// 	}
// }

func TestMaxNumberOfFiles(t *testing.T) {
	service := New()
	service.Start()

	service.SetMaxProcessingFiles(1)
	assert.Equal(t, 1, service.GetCurrentProcessingFiles())

	longTimeTask := testProcessMultipleCSVs(service, []string{"7e4fe5cb-d40a-4f7a-8fe4-374091f19a47"}, true)
	quickTask := testProcessMultipleCSVs(service, []string{"479bb757-8a7a-4e4e-9b71-845192422579"}, false)
	fmt.Println("longTimeTask: ", len(longTimeTask))
	fmt.Println("quickTask: ", len(quickTask))

	var longTaskExecutedBeforeQuick bool
checkresults:
	for {
		select {
		case <-longTimeTask[0]:
			fmt.Println("123 ---------------------------------")
			longTaskExecutedBeforeQuick = true
		case <-quickTask[0]:
			fmt.Println("456 ---------------------------------")
			if !longTaskExecutedBeforeQuick {
				assert.Fail(t, "long task should have executed before short due to limit")
			}
			break checkresults
		default:
			assert.Less(t, service.GetCurrentProcessingFiles(), 2)
		}
	}

	assert.True(t, true)
}

func testProcessMultipleCSVs(service *CSVService, requestIDs []string, longDelay bool) []chan error {
	handler := alterWithShortDelay
	if longDelay {
		handler = alterWithLongDelay
	}

	resultChans := make([]chan error, 0, len(requestIDs))
	for i, id := range requestIDs {
		c := service.ProcessCSV(
			context.Background(),
			id,
			"test_game_data.csv",
			fmt.Sprintf("add_words_intensive_work_to_every_field_%d.csv", i),
			handler,
		)
		resultChans = append(resultChans, c)
	}
	return resultChans
}

func wrapTestLog(a ...interface{}) {
	var text []interface{}

	pre := "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n"
	post := "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n "
	label := "[TEST]: "
	text = append(text, pre)
	text = append(text, label)
	text = append(text, a...)
	text = append(text, post)
	fmt.Println(text...)

}

// func GetCSVRecords(filePath string) {
// 	// Open the file
// 	csvfile, err := os.Open(filePath)
// 	if err != nil {
// 		log.Fatalln("Couldn't open the csv file", err)
// 	}

// 	// Parse the file
// 	r := csv.NewReader(csvfile)
// 	//r := csv.NewReader(bufio.NewReader(csvfile))

// 	// Iterate through the records
// 	fmt.Println("<<<<<<<< ALTERED CSV CONTENTS >>>>>>>>")
// 	for {
// 		// Read each record from csv
// 		record, err := r.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Printf("Record %s\n", record)
// 	}
// 	fmt.Println(">>>>>>>>> ALTERED CSV CONTENTS END <<<<<<<<<<")

// }
