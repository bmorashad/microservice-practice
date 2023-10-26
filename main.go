package main

import (
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		http.ListenAndServe(":8011", nil)
	}()
	a := App{}
	// a.Initializer( "postgres", "123", "localhost", "postgres")
	// a.Initializer( "postgres", "123", "server_postgres", "postgres")
	// a.Initializer( "postgres", "123", "localhost", "postgres")
	a.Initializer( "root", "root", "localhost", "postgres")

	a.Run(":8010")
}
