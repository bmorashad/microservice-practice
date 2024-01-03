package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"practice-server/tracing"

	"github.com/joho/godotenv"
)

func main() {
	serviceName := "random-product-generator"
	ctx := context.Background()
	// shutdown, err = tracing.InitZipkinOtelTrace(ctx, serviceName, "products", "dev")
	shutdown, err := tracing.InitJaegerOtelTrace(ctx, serviceName, "products", "dev")
	if err != nil {
		log.Fatalf("failed to initialize stdouttrace export pipeline: %v", err)
	}
	// shutting down causes otel middleware to stop working
	defer func() {
		fmt.Println("Shutting down")
		shutdown(ctx)
	}()
	go func() {
		pprofPort := fmt.Sprintf(":%s", os.Getenv("PPROF_PORT"))
		http.ListenAndServe(pprofPort, nil)
	}()
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	a := App{}
	a.Initializer(serviceName)
	a.Run()
}
