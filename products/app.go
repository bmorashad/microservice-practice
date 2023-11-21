package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
  "sync"

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
	m := NewMetrics(reg)
	m.concurrentExecutions.Set(2)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	a.Router = mux.NewRouter()

	a.initializeRoutes(promHandler)
  initDB(a.DB)
  go productToMerchantBatchProcess(a.DB)
}

func initDB(db *sql.DB) {
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

func (a *App) Run() {
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
  // if count > 10 || count < 1 {
  if count < 1 {
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

func logProductCount(db *sql.DB) {
  uptimeTicker := time.NewTicker(10 * time.Second)
  for {
    select {
    case <- uptimeTicker.C:
      count, err := countProducts(db)
      if err != nil {
        log.Println("Error occurred while counting products:", err)
      }
      log.Println(count, "products are in the table")
    }
  }
}

func productToMerchantBatchProcess(db *sql.DB) {
  uptimeTicker := time.NewTicker(5 * time.Second)
  for {
    select {
    case <- uptimeTicker.C:
      sendProductsToMerchant(db)
    }
  }
}

func sendProductsToMerchant(db *sql.DB) {
  products, err := getProducts(db, lastSentProductID, 20)
  if err != nil {
    log.Println(err)
  }
  if len(products) > 0 {
    lastProductId := products[len(products) - 1].ID
    firstProductId := products[0].ID
    if lastProductId == lastSentProductID {
      return
    }
    err = _sendToMerchant(products)
    if err != nil {
      log.Println("Error occurred while sending products to merchant:", err)
      return
    }
    lastSentProductID = lastSentProductID+len(products)
    log.Println(len(products),  "products sent to merchant -", "from ID:", firstProductId, "to ID:", lastProductId, "Last sent product ID:", lastSentProductID)
  }
}

func _sendToMerchant(products []product) error {
  merchantUrl := fmt.Sprintf("http://%s:%s/receive/products", os.Getenv("MERCHANT_SERVICE_HOST"), os.Getenv("MERCHANT_SERVICE_PORT"))
  jsonProducts, err := json.Marshal(products)
  if err != nil {
    return err
  }
  resp, err := http.Post(merchantUrl, "application/json", bytes.NewBuffer(jsonProducts))
  if err != nil {
    return err
  }
  defer resp.Body.Close()
  return nil
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

func (a *App) callProductsApi() error {
  products, err := getProducts(a.DB, 0, 10)
  if err != nil {
    log.Println("Error occurred while calling products api:", err)
    return err
  }
  lastProductCount := 1
  if len(products) > 0 {
    lastProductCount = products[len(products)-1].ID
  }
  for i := lastProductCount; i < 1; i++ {}
  return nil
}

func (a *App) createRandomProducts(w http.ResponseWriter, r *http.Request) {
  err := a.callProductsApi()
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, "Internal server error")
    return
  }
  numOfRandProductsToCreate := 10
  var wg sync.WaitGroup
  // wg.Add(numOfRandProductsToCreate)
  var mu sync.Mutex
  createdProductCount := 0
  for i := 0; i < numOfRandProductsToCreate; i++ {
    go _createRandomProducts(a.DB, &createdProductCount, &wg, &mu)
  }
  // wg.Wait()
  // ch := make(chan bool)
  // go func() {
  //   {
  //     <-ch
  //     log.Println("Successfully created product count:", createdProductCount)
  //     createdProductCount = 0
  //     close(ch)
  //   }
  // }()
  respondWithJSON(w, http.StatusOK, map[string]string{"result": "started random products creation"})
}

func _createRandomProducts(db *sql.DB, cpc *int, wg *sync.WaitGroup, mu *sync.Mutex) {
  // defer wg.Done()
  randomProductServiceHost := fmt.Sprintf("%s", os.Getenv("RANDOM_PRODUCT_INFO_SERVICE_HOST"))
  randomProductServicePort := fmt.Sprintf("%s", os.Getenv("RANDOM_PRODUCT_INFO_SERVICE_PORT"))
  var product = &product{}
  response, err := http.Get(fmt.Sprintf("http://%s:%s/random-product/info", randomProductServiceHost, randomProductServicePort))
  if err != nil {
    log.Println("Error occurred while calling random-product-info: ", err)
    return
  }
  if response.Status != "200" {
    response, err = http.Get(fmt.Sprintf("http://%s:%s/random-product/info", randomProductServiceHost, randomProductServicePort))
    if err != nil {
      log.Println("Error while calling random-product: ", err)
      log.Println("Here is the response", response)
      return
    }
  }
  err = json.NewDecoder(response.Body).Decode(&product)

  if err != nil {
    log.Println("Error while reading response body:", err)
    log.Println("Here is the response", response)
    return
  }

  _, err = product.createProduct(db)
  if err != nil {
    log.Println("Error while creating random product:", err)
  }

  // mu.Lock()
  // *cpc += 1
  // mu.Unlock()
  
  // log.Println("Created random product:", product)
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
