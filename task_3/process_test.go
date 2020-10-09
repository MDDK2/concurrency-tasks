package task_3

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestJoin(t *testing.T) {
	service := New()
	service.Start()

	resultChan1 := service.ProcessCSV(
		context.Background(),
		"test_game_data.csv",
		func(ctx context.Context, record []string) error {
			// very intensive work
			time.Sleep(time.Second * 5)
			return nil
		})

	resultChan2 := service.ProcessCSV(
		context.Background(),
		"test_game_data.csv",
		func(ctx context.Context, record []string) error {
			// work
			time.Sleep(time.Millisecond * 5)
			return nil
			// return errors.New()
		})

	//time.Sleep(time.Second * 2)
	// GetCSVRecords("test_game_data.csv")
	results := make([]error, 0, 2)
	for {
		if len(results) == 2 {
			break
		}

		select {
		case r1 := <-resultChan1:
			fmt.Println("r1: ", r1)
			results = append(results, r1)
		case r2 := <-resultChan2:
			fmt.Println("r2: ", r2)
			results = append(results, r2)
		default:
		}

	}

	service.Stop()

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
