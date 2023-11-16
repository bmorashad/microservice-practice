package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

var lastSentProductID = 0

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
  go productToMerchantBatchProcess(a.DB)
}

func initDB(db *sql.DB) {
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
  serverPort := fmt.Sprintf(":%s", os.Getenv("PRODUCTS_SERVICE_PORT"))
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
  products, err := getProducts(a.DB, start, count)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }
  respondWithJSON(w, http.StatusOK, products)
}

func (a *App) getAllProducts(w http.ResponseWriter, r *http.Request) {
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

func productToMerchantBatchProcess(db *sql.DB) {
  uptimeTicker := time.NewTicker(3 * time.Second)
  for {
    select {
    case <- uptimeTicker.C:
      sendProductsToMerchant(db)
    }
  }
}

func sendProductsToMerchant(db *sql.DB) {
  fmt.Println("Sending product to merchant")
  products, err := getProducts(db, lastSentProductID, lastSentProductID+10)
  if err != nil {
    fmt.Println(err)
  }
  fmt.Println("Len of products", len(products))
  if len(products) > 0 {
    lastProductId := products[len(products) - 1].ID
    firstProductId := products[0].ID
    fmt.Println("Last Product ID:", lastProductId)
    fmt.Println("First Product ID:", firstProductId)
    fmt.Println("Last Sent Product ID:",lastSentProductID)
    if lastProductId == lastSentProductID {
      return
    }
    _sendToMerchant(products)
    lastSentProductID = lastSentProductID+len(products)
  }
}

func _sendToMerchant(products []product) {
  merchantUrl := fmt.Sprintf("http://%s:%s/receive/products", os.Getenv("MERCHANT_SERVICE_HOST"), os.Getenv("MERCHANT_SERVICE_PORT"))
  jsonProducts, err := json.Marshal(products)
  if err != nil {
    fmt.Println("Error marshalling JSON: ", err)
    return
  }
  resp, err := http.Post(merchantUrl, "application/json", bytes.NewBuffer(jsonProducts))
  if err != nil {
    fmt.Println("Error occurred while sending ", err)
    return
  }
  defer resp.Body.Close()
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

func (a *App) createRandomProducts(w http.ResponseWriter, r *http.Request) {
  products, err := getProducts(a.DB, 0, 10)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, "Internal server error")
    return
  }
  lastProductCount := 1
  if len(products) > 0 {
    lastProductCount = products[len(products)-1].ID
  }
  for i := lastProductCount; i < 1; i++ {}
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }
  for i := 0; i < 10; i++ {
    go func() {
      randomProductServiceHost := fmt.Sprintf("%s", os.Getenv("RANDOM_PRODUCT_INFO_SERVICE_HOST"))
      randomProductServicePort := fmt.Sprintf("%s", os.Getenv("RANDOM_PRODUCT_INFO_SERVICE_PORT"))
      var product = &product{}
      response, err := http.Get(fmt.Sprintf("http://%s:%s/random-product/info", randomProductServiceHost, randomProductServicePort))
      if err != nil {
        fmt.Println(err)
        return
      }
      if response.Status != "200" {
        response, err = http.Get(fmt.Sprintf("http://%s:%s/random-product/info", randomProductServiceHost, randomProductServicePort))
        if err != nil {
          fmt.Println(err)
          return
        }
      }
      err = json.NewDecoder(response.Body).Decode(&product)
      // body, err := ioutil.ReadAll(response.Body)
      // fmt.Println("Response Status:", response)
      // fmt.Println("Response Status:", response.Status)
      // fmt.Println("Response Body:", string(body))

      if err != nil {
        log.Fatal("Error reading response body:", err)
      }

      fmt.Println("This is product", product)
      _, err = product.createProduct(a.DB)
      if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
      }
    }()
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

func (a *App) errorTest(w http.ResponseWriter, r *http.Request) {
  // respondWithJSON(w, http.StatusForbidden, map[string]string{"result": "hello-world!"})
  respondWithError(w, http.StatusForbidden, "Error occurred")
}
func (a *App) truncate(w http.ResponseWriter, r *http.Request) {
  err := truncate(a.DB)
  if err != nil {
    respondWithError(w, http.StatusConflict, "Error occurred while truncating products table")
  }
  respondWithJSON(w, http.StatusOK, "successfully truncated products table");
}

func (a *App) initializeRoutes(http.Handler) {
  a.Router.HandleFunc("/products/all", a.getAllProducts).Methods("GET")
  a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
  a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
  a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
  a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
  a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
  a.Router.HandleFunc("/create-products/random", a.createRandomProducts).Methods("GET")
  a.Router.HandleFunc("/products/truncate", a.truncate).Methods("GET")
  a.Router.HandleFunc("/ping", a.ping).Methods("GET")
  a.Router.HandleFunc("/error", a.errorTest).Methods("GET")
  // a.Router.HandleFunc("/healthz", a.ping).Methods("GET")
  // a.Router.HandleFunc("/", a.ping).Methods("GET")

  // a.Router.Handle("/metrics", promhttp.Handler())
}
