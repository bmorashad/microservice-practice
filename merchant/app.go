package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

var lastSentProductID = 0

func (a *App) Initializer(user, password, host, port, dbname, serviceName string) {
	cfg := mysql.Config{
		Net:    "tcp",
		User:   user,
		Passwd: password,
		Addr:   fmt.Sprintf("%s:%s", host, port),
		DBName: dbname,
	}

	log.Println(cfg.FormatDSN())

	var err error
	a.DB, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Println("Error occurred while connecting to db:> ", err.Error())
		log.Fatal(err)
	}

	reg := prometheus.NewRegistry()
	// m := NewMetrics(reg)
	// m.concurrentExecutions.Set(2)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	a.Router = mux.NewRouter()
	a.Router.Use(otelmux.Middleware(serviceName))

	a.initializeRoutes(promHandler)
	initDB(a.DB)
}

func initDB(db *sql.DB) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	res, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS merchantproducts (id int not null auto_increment, name varchar(255), price varchar(255), PRIMARY KEY (id))")
	if err != nil {
		log.Fatal("Error when creating merchantproducts table: ", err)
	}
	no, err := res.RowsAffected()
	if err != nil {
		log.Fatal("Error when fetching rows: ", err)
		return
	}
	log.Printf("rows affected: %d\n", no)
}

func (a *App) Run() {
	serverPort := fmt.Sprintf(":%s", os.Getenv("MERCHANT_SERVICE_PORT"))
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

func (a *App) receiveProducts(w http.ResponseWriter, r *http.Request) {
	var products []product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&products); err != nil {
		log.Println("Error corred while decoding received product:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid payload")
		return
	}
	defer r.Body.Close()

	for _, p := range products {
		if _, err := p.createProduct(a.DB); err != nil {
			log.Println("Error corred while storing merchant product:", err)
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	log.Println("Successfully stored received product")
	respondWithJSON(w, http.StatusCreated, products)
}

func (a *App) ping(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}
	products, err := getProducts(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, products)
}

func (a *App) reset(w http.ResponseWriter, r *http.Request) {
	lastSentProductID = 0
	err := truncate(a.DB)
	if err != nil {
		log.Println("Error while truncating db for application reset", err)
		respondWithError(w, http.StatusConflict, "Error occurred while truncating products table")
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "application reset successfull"})
}

func (a *App) initializeRoutes(http.Handler) {
	a.Router.HandleFunc("/reset", a.reset).Methods("GET")
	a.Router.HandleFunc("/receive/products", a.receiveProducts).Methods("POST")
	a.Router.HandleFunc("/ping", a.ping).Methods("GET")
	a.Router.HandleFunc("/merchant/products", a.getProducts).Methods("GET")
	// a.Router.Handle("/metrics", promhttp.Handler())
}
