package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	go func() {
    pprofPort := fmt.Sprintf(":%s", os.Getenv("PPROF_PORT"))
    http.ListenAndServe(pprofPort, nil)
  }()
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }
  a := App{}
  a.Initializer(os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
  a.Run(":8010")
}
