// main.go

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type InputData struct {
	ToSort [][]int `json:"to_sort"`
}

type ResponseData struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       int64   `json:"time_ns"`
}

func processSingle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request", r.Method)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	log.Printf("Received request body: %s", body)

	var inputData InputData
	err = json.Unmarshal(body, &inputData)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := make([][]int, len(inputData.ToSort))

	for i, arr := range inputData.ToSort {
		sortedArray := make([]int, len(arr))
		copy(sortedArray, arr)
		sort.Ints(sortedArray)
		sortedArrays[i] = sortedArray
	}

	response := ResponseData{
		SortedArrays: sortedArrays,
		TimeNS:       time.Since(startTime).Nanoseconds(),
	}

	sendJSONResponse(w, response)
}

func processConcurrent(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request", r.Method)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	log.Printf("Received request body: %s", body)

	var inputData InputData
	err = json.Unmarshal(body, &inputData)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	var wg sync.WaitGroup
	sortedArrays := make([][]int, len(inputData.ToSort))
	mutex := &sync.Mutex{}

	for i, arr := range inputData.ToSort {
		wg.Add(1)
		go func(index int, array []int) {
			defer wg.Done()
			sortedArray := make([]int, len(array))
			copy(sortedArray, array)
			sort.Ints(sortedArray)

			mutex.Lock()
			sortedArrays[index] = sortedArray
			mutex.Unlock()
		}(i, arr)
	}

	wg.Wait()

	response := ResponseData{
		SortedArrays: sortedArrays,
		TimeNS:       time.Since(startTime).Nanoseconds(),
	}

	sendJSONResponse(w, response)
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/process-single", processSingle)
	http.HandleFunc("/process-concurrent", processConcurrent)

	http.ListenAndServe(":8000", nil)
}
