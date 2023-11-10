package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/danyaizm/orders-api/models"
	"github.com/danyaizm/orders-api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Order struct {
	Repo storage.OrderRepo
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CurstomerID uuid.UUID         `json:"customer_id"`
		LineItems   []models.LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()

	order := models.Order{
		ID:         rand.Uint64(),
		CustomerID: body.CurstomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	if err := o.Repo.Insert(r.Context(), order); err != nil {
		fmt.Println("failed to create: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(order)
	if err != nil {
		fmt.Println("failed to marshall object: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
	w.WriteHeader(http.StatusCreated)
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	cursor, err := strconv.ParseUint(cursorStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := o.Repo.FindAll(r.Context(), storage.FindAllPage{
		Size:   50,
		Offset: cursor,
	})
	if err != nil {
		fmt.Println("failed to find all orders: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []models.Order `json:"items"`
		Next  uint64         `json:"next,omitempty"`
	}
	response.Items = res.Orders
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal response data: ", err)
	}

	w.Write(data)
	w.WriteHeader(http.StatusOK)
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := o.Repo.FindByID(r.Context(), id)
	if errors.Is(err, storage.ErrorNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find an order: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(order); err != nil {
		fmt.Println("failed to marshall object: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := o.Repo.FindByID(r.Context(), id)
	if errors.Is(err, storage.ErrorNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find order by id in order to update: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()

	var response struct {
		Error     string `json:"error"`
		ErrorCode int    `json:"error_code"`
	}

	switch body.Status {
	case shippedStatus:
		if order.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Error = "order is already shipped"
			response.ErrorCode = ErrorAlreadyShippedCode
			json.NewEncoder(w).Encode(response)
			return
		}
		order.ShippedAt = &now
	case completedStatus:
		if order.CompletedAt != nil || order.ShippedAt == nil {
			json.NewEncoder(w).Encode(response)
			response.Error = "order is already completed or not shipped"
			response.ErrorCode = ErrorNotShippedOrAlreadyCompletedCode
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		order.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		response.Error = "bad status string"
		response.ErrorCode = ErrorBadStatusStringCode
		json.NewEncoder(w).Encode(response)
		return
	}

	if err := o.Repo.Update(r.Context(), order); err != nil {
		fmt.Println("failed to update order: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(order); err != nil {
		fmt.Println("failed to encode an order in update case: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := o.Repo.DeleteByID(r.Context(), id); errors.Is(err, storage.ErrorNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to delete an order: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
