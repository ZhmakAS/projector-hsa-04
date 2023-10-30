package main

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
)

type product struct {
	Title       string
	Price       float32
	Description string
	Image       string
}

type productsReturn struct {
	Products []product
	Title    string
}

func getProducts() ([]product, error) {
	response, _ := http.Get("https://fakestoreapi.com/products?limit=5")
	bytes, _ := io.ReadAll(response.Body)
	response.Body.Close()

	var productsList []product

	err := json.Unmarshal(bytes, &productsList)

	if err != nil {
		return nil, err
	}

	return productsList, nil
}

func productsHandler(response http.ResponseWriter, request *http.Request) {
	products, err := getProducts()

	if err != nil {
		panic(err)
	}

	view, _ := template.ParseFiles("./views/products.html")
	data := productsReturn{Products: products, Title: "Products"}

	view.Execute(response, data)
}

func main() {
	styles := http.FileServer(http.Dir("./views/stylesheets"))
	http.Handle("/styles/", http.StripPrefix("/styles/", styles))
	http.HandleFunc("/products", productsHandler)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		panic(err)
	}
}
