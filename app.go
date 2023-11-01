package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"encoding/json"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initializer(user, password, host, port, dbname string) {
  // connectionString := fmt.Sprintf("postgres://%s:%s@server_postgres:5432/%s?sslmode=disable", user, password, dbname)
  // connectionString := fmt.Sprintf("user=%s password=%s  dbname=%s sslmode=disable", user, password, dbname)

   cfg := mysql.Config{
        // User:   os.Getenv("DBUSER"),
        // Passwd: os.Getenv("DBPASS"),
        // User:   user,
        // Passwd: password,
        // Net:    "tcp",
        // Addr:   "127.0.0.1:3306",
        // Addr:   "localhost:3306",
        // Addr:   "db:3306",
        Net:    "tcp",
        User:   user,
        Passwd: password,
        // Addr:   "db:3306",
        // Addr:   "localhost:3307",
        Addr:   fmt.Sprintf("%s:%s", host, port),
        DBName: dbname,
    }

  fmt.Println(cfg.FormatDSN())

	var err error
	a.DB, err = sql.Open("mysql", cfg.FormatDSN())
  // connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4", user, password, host, port, dbname)
  // connectionString := fmt.Sprintf("root:root@tcp(docker.for.mac.localhost:3037)/ecommerce?charset=utf8mb4")
  // connectionString := fmt.Sprintf("root:root@tcp(localhost:3036)/ecommerce?charset=utf8mb4")
  // connectionString := fmt.Sprintf("root:root@tcp(localhost:3037)/ecommerce?allowNativePasswords=false&checkConnLiveness=false&maxAllowedPacket=0")
  // connectionString := fmt.Sprintf("root:root@tcp(localhost:3037)/ecommerce?allowNativePasswords=false&checkConnLiveness=false&maxAllowedPacket=0")

  // fmt.Println("Connection String:> ", connectionString)
  // a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
    fmt.Println("Error occurred while connecting to db:> ", err.Error())
		log.Fatal(err)
	}

	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	m.concurrentExecutions.Set(2)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	a.Router = mux.NewRouter()

	a.initializeRoutes(promHandler)
  initDB(a.DB)
}

func initDB(db *sql.DB) {
  log.Printf("Hello there I'm working")  
  log.Printf("Hello there I'm working")  
  log.Printf(".")  
  log.Printf("..")  
  log.Printf("...")  
  // ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
  // defer cancelFunc()
  // res, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS ecommerce")  
  // if err != nil {  
  //   log.Printf("Error %s when creating DB\n", err)
  //   return
  // }
  // no, err := res.RowsAffected()  
  // if err != nil {  
  //   log.Printf("Error %s when fetching rows", err)
  //   return
  // }
  // log.Printf("rows affected: %d\n", no)  
  ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancelFunc()
  res, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS products (id int not null auto_increment, name varchar(255), price varchar(255), PRIMARY KEY (id))")  
  if err != nil {
    log.Fatal("Error when creating products table: ", err)
  }
  no, err := res.RowsAffected()  
  if err != nil {  
    log.Fatal("Error when fetching rows: ", err)
    return
  }
  log.Printf("rows affected: %d\n", no)  
}

func (a *App) Run(addr string) {
  serverPort := fmt.Sprintf(":%s", os.Getenv("SERVER_PORT"))
  log.Fatal(http.ListenAndServe(serverPort, a.Router))

}

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid product Id")
    return
  }
  p := product{ID: id}
  if err := p.getProduct(a.DB); err != nil {
    switch err {
    case sql.ErrNoRows:
      respondWithError(w, http.StatusNotFound, "Prodcut not found")
    default:
      respondWithError(w, http.StatusInternalServerError, err.Error())
    }
    return
  }
  respondWithJSON(w, http.StatusOK, p)
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
func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
  count, _ := strconv.Atoi(r.FormValue("count"))
  start, _ := strconv.Atoi(r.FormValue("start"))
  if count > 10 || count < 1 {
    count = 10
  }
  if start < 0 {
    start = 0
  }
  products, err := getAllProducts(a.DB)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }
  respondWithJSON(w, http.StatusOK, products)
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
  var p product
  decoder := json.NewDecoder(r.Body)
  if err := decoder.Decode(&p); err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid payload")
    return
  }
  defer r.Body.Close()

  if _, err := p.createProduct(a.DB); err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }
  respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid product ID")
    return
  }

  var p product
  decoder := json.NewDecoder(r.Body)
  if err := decoder.Decode(&p); err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
    return
  }
  defer r.Body.Close()
  p.ID = id

  if err := p.updateProduct(a.DB); err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
    return
  }

  p := product{ID: id}
  if err := p.deleteProduct(a.DB); err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) longRunningProcess(w http.ResponseWriter, r *http.Request) {
  var productPrefix string
  productPrefix = "Product"
  products, err := getAllProducts(a.DB)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, "Internal server error")
    return
  }
  lastProductCountInt := 1
  if len(products) > 0 {
    lastProduct := products[len(products)-1]
    lastProductCount := strings.TrimSpace(lastProduct.Name[len(productPrefix):len(lastProduct.Name)])
    lastProductCountInt, err = strconv.Atoi(lastProductCount)
  }
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }
  for i := lastProductCountInt + 1; i < lastProductCountInt+1+10; i++ {
    p := &product{
      Name:  fmt.Sprintf(productPrefix+"%s", fmt.Sprint(i)),
      Price: 200,
    }
    _, err := p.createProduct(a.DB)
    if err != nil {
      respondWithError(w, http.StatusInternalServerError, err.Error())
      return
    }
  }
  go func() error {
    response, err := http.Get("https://pokeapi.co/api/v2/pokedex/kanto/")
    if err != nil {
      fmt.Println(err.Error())
      return err
    }
    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
      fmt.Println(err.Error())
      return err
    }
    fmt.Println(string(responseData[0]))
    return nil
  }()
  respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) ping(w http.ResponseWriter, r *http.Request) {
  respondWithJSON(w, http.StatusOK, map[string]string{"result": "hello-world!"})
}

func (a *App) initializeRoutes(http.Handler) {
  a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
  a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
  a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
  a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
  a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
  a.Router.HandleFunc("/create-products/random", a.longRunningProcess).Methods("GET")
  a.Router.HandleFunc("/ping", a.ping).Methods("GET")

  // a.Router.Handle("/metrics", promhttp.Handler())
}
