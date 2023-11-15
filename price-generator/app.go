package main

import (
	"math/rand"
	"database/sql"
	"fmt"
	"log"
	"os"

	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initializer(user, password, host, port, dbname string) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	m.concurrentExecutions.Set(2)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	a.Router = mux.NewRouter()
	a.initializeRoutes(promHandler)
}


func (a *App) Run(addr string) {
  serverPort := fmt.Sprintf(":%s", os.Getenv("SERVER_PORT"))
  log.Fatal(http.ListenAndServe(serverPort, a.Router))

}

func respondWithError(w http.ResponseWriter, code int, message string) {
  respondWithJSON(w, code, map[string]string{"error": message})
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
  response, _ := json.Marshal(payload)
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(code)
  w.Write(response)
}

func (a *App) generateRandomPrice(w http.ResponseWriter, r *http.Request) {
  price := rand.Intn(1000) + 400
  respondWithJSON(w, http.StatusCreated, price)
}

func (a *App) ping(w http.ResponseWriter, r *http.Request) {
  respondWithJSON(w, http.StatusOK, map[string]string{"result": "hello-world!"})
}

func (a *App) initializeRoutes(http.Handler) {
  a.Router.HandleFunc("/random-price/generate", a.generateRandomPrice).Methods("GET")
  a.Router.HandleFunc("/ping", a.ping).Methods("GET")
  // a.Router.Handle("/metrics", promhttp.Handler())
}
