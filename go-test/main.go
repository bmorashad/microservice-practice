package main

import (
	"fmt"
	"strconv"
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
  fmt.Println(count, start)


}
