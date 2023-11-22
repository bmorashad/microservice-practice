package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ErrLog struct {
  message string
  err error
  resp *http.Response
  count int
}

func test() {
  var wg sync.WaitGroup
  wg.Add(10)
  errChan := make(chan ErrLog, 10)
  errLog := make(map[string]ErrLog)

  for i := 0; i < 10; i++ {
    go _test(&wg, errChan)
  }
  go func() {
    wg.Wait()
    close(errChan)
    for log := range errChan {
      if value, exists := errLog[log.message]; exists {
        log.count = value.count + 1
      } else {
        log.count = 1
      }
      errLog[log.message] = log
      // Perform other operations with errLog
    }
    fmt.Println("Error Map")
    fmt.Println("---")
    mapPrint(errLog)

  }()
}

func mapPrint(myMap map[string]ErrLog) {
  for key, value := range myMap {
    if value.resp != nil {
      fmt.Printf("%s: \t%s, %s, %d, %d, %s, %s\n", key, value.message, value.err, value.count, value.resp.StatusCode, value.resp.Status, value.resp.Header)
    } else {
      fmt.Printf("%s: \t%s, %s, %d, <nil>\n", key, value.message, value.err, value.count)
    }
  }
}

func callFakeApi() (*http.Response, error) {
  url := "https://jsonplaceholder.typicode.com/posts/1"
	// Make a GET request to the fake API
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return &http.Response{}, err
	}
	defer resp.Body.Close()
  return resp, nil
}

func _test(wg *sync.WaitGroup, errCh chan ErrLog) {
  defer wg.Done()
  rand := rand.Intn(4)
  if rand < 2 {
    // fmt.Println("< 2")
    // log.Println("Rand doesn't match to create product: randNumber: ", rand)
    errLog := ErrLog{
      message: fmt.Sprintf("Randoment Error"),
      err: fmt.Errorf("MyError"),
      resp: nil,
    }
    errCh <- errLog
    // log.Println("Error occurred while calling random-product-info: ", err)
    return
  }
  resp, err := callFakeApi()
  if err != nil {
    errLog := ErrLog{
      message: "CallApi Error",
      err: fmt.Errorf("MyAPIError"),
      resp: nil,
    }
    errCh <- errLog
    return
  }
  // fmt.Println(">= 2")
  errLog := ErrLog{
    message: "Different Error",
    err: fmt.Errorf("MyDifferentError"),
    resp: resp,
  }
  errCh <- errLog
  // log.Println("Error occurred while calling random-product-info: ", err)
  return
}

func main() {
  test()
  time.Sleep(65 * time.Millisecond)
  fmt.Println("==================")

  countParam := ""
  startParam := ""
  count, _ := strconv.Atoi(countParam)
  start, _ := strconv.Atoi(startParam)
  fmt.Println(count, start)
  if count > 10 || count < 1 {
    count = 10
  }
  if start < 0 {
    start = 0
  }
  go func() {
    log.Println("Im gonna be Fatal")
    // time.Sleep(5 * time.Second) 
    log.Fatal("I already Fataled")
  }()
  fmt.Println(count, start)
  fmt.Println("Random Test")
  names := []string{"Milk", "Ice Cream", "Yourgut", "Car", "Van", "T-Shirt", "Cookies"}
  rand.Seed(time.Now().Unix())
  pickedName := names[rand.Intn(len(names))]
  fmt.Println("Picked Name:", pickedName)
  log.Println("Picked Name:", pickedName)
  // time.Sleep(8 * time.Second) 
  log.Println("Sleep over...")
  log.Println("Print random int", rand.Intn(2))

}
