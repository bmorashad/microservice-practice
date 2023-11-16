package main

import (
	"database/sql"
	"fmt"
	"log"
)

type product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price int `json:"price"`
}


func (p *product) getProduct(db *sql.DB) error {
  return db.QueryRow("SELECT name, price FROM products WHERE id=?", p.ID).Scan(&p.Name, &p.Price)
}
func (p *product) updateProduct(db *sql.DB) error {
  _, err := db.Exec("UPDATE products SET name=?, price=? WHERE id=?", p.Name, p.Price, p.ID)
  return err
}
func (p *product) createProduct(db *sql.DB) (sql.Result, error) {
  return db.Exec("INSERT INTO products(name, price) VALUES(?, ?)", p.Name, p.Price)
}
func (p *product) deleteProduct(db *sql.DB) error {
  _, err := db.Exec("DELETE FROM products WHERE id=?", p.ID)
  return err
}

func truncate(db *sql.DB) error {
  _, err := db.Exec("TRUNCATE TABLE products")
  if err != nil {
    log.Fatal(err)
    return err
  }
  return nil
}

func getAllProducts(db *sql.DB) ([]product, error) {
  rows, err := db.Query("SELECT * from products ORDER BY id")
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  products := []product{}
  for rows.Next() {
    var p product
    if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
      return nil, err
    }
    products = append(products, p)
  }
  return products, nil
}
func getProducts(db *sql.DB, start, count int) ([]product, error) {
  fmt.Println("Start & Count: ", start, count)
  rows, err := db.Query(
    "SELECT id, name, price FROM products LIMIT ? OFFSET ?", 
    count, start)
    if err != nil {
      return nil, err
    } 
    defer rows.Close()
    products := []product{}
    for rows.Next() {
      var p product
      if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
        return nil, err
      }
      products = append(products, p)
    }
    return products, nil

  }