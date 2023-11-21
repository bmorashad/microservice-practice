package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	// "time"

	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
)

type FakeNameResponseData struct {
    Results []struct {
        Name struct {
            First string `json:"first"`
            Last  string `json:"last"`
            Title string `json:"title"`
        } `json:"name"`
    } `json:"results"`
}

type App struct {
	Router *mux.Router
}

func (a *App) Initializer() {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	m.concurrentExecutions.Set(2)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	a.Router = mux.NewRouter()
	a.initializeRoutes(promHandler)
}


func (a *App) Run() {
  serverPort := fmt.Sprintf(":%s", os.Getenv("RANDOM_PRODUCT_INFO_SERVICE_PORT"))
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

func (a *App) generateRandomProductInfoFromList() (product, error) {
  names := []string{"Milk", "Ice Cream", "Yourgut", "Car", "Van", "T-Shirt", "Cookies"}
  // rand.Seed(time.Now().Unix())
  pickedName := names[rand.Intn(len(names))]
  var p = product{}
  p.Name = pickedName
  price, err := generatePrice()
  if err != nil {
    return product{}, err
  }
  p.Price = price
  // p.Price = 400
  return p, nil
}

func (a *App) generateRandomProductInfoFromFakeApi() (product, error) {
  var p = product{}
  response, err := http.Get("https://randomuser.me/api/?results=1")
  if err != nil {
    fmt.Println(err.Error())
    return product{}, err
  }
  responseData := FakeNameResponseData{}
  err = json.NewDecoder(response.Body).Decode(&responseData)
  if err != nil {
    fmt.Println(err.Error())
    return product{}, err
  }
  firstName := responseData.Results[0].Name.First
  lastName := responseData.Results[0].Name.Last
  fullName := firstName + lastName 
  p.Name = fullName
  price, err := generatePrice()
  if err != nil {
    return product{}, err
  }
  p.Price = price
  // p.Price = 400
  return p, nil
}

func generatePrice() (int, error) {
  priceGeneratorServiceHost := fmt.Sprintf("%s", os.Getenv("PRICE_SERVICE_HOST"))
  priceGeneratorServicePort := fmt.Sprintf("%s", os.Getenv("PRICE_SERVICE_PORT"))
  var price int 
  response, err := http.Get(fmt.Sprintf("http://%s:%s/random-price/generate", priceGeneratorServiceHost, priceGeneratorServicePort))
  if err != nil {
    return -1, err
  }
  err = json.NewDecoder(response.Body).Decode(&price)
  if err != nil {
    return -1, err
  }
  return price, nil
}

func (a *App) generateRandomProduct(w http.ResponseWriter, r *http.Request) {
  p, err := a.generateRandomProductInfoFromList()
  if err != nil {
    log.Println(err)
    respondWithError(w, http.StatusInternalServerError, "error occurred while generating product")
  }
  respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) ping(w http.ResponseWriter, r *http.Request) {
  respondWithJSON(w, http.StatusOK, map[string]string{"result": "hello-world!"})
}

func (a *App) initializeRoutes(http.Handler) {
  a.Router.HandleFunc("/random-product/info", a.generateRandomProduct).Methods("GET")
  a.Router.HandleFunc("/ping", a.ping).Methods("GET")
  // a.Router.Handle("/metrics", promhttp.Handler())
}
