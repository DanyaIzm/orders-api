package handlers

import (
	"net/http"
)

type Order struct{}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create a new order"))
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("List all orders"))
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get an order by ID"))
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Update an order by ID"))
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Delete an order by ID"))
}
