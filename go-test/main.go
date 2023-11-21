package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func main() {
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
    time.Sleep(5 * time.Second) 
    log.Fatal("I already Fataled")
  }()
  fmt.Println(count, start)
  fmt.Println("Random Test")
  names := []string{"Milk", "Ice Cream", "Yourgut", "Car", "Van", "T-Shirt", "Cookies"}
  rand.Seed(time.Now().Unix())
  pickedName := names[rand.Intn(len(names))]
  fmt.Println("Picked Name:", pickedName)
  log.Println("Picked Name:", pickedName)
  time.Sleep(8 * time.Second) 
  log.Println("Sleep over...")
  

  
}
