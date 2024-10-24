package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type OrderStatus string

const (
	StatusProses  OrderStatus = "Diproses"
	StatusAntar   OrderStatus = "Diantar"
	StatusSelesai OrderStatus = "Selesai"
)

type MenuItem struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Order struct {
	ID     int         `json:"id"`
	Items  []MenuItem  `json:"items"`
	Total  float64     `json:"total"`
	Status OrderStatus `json:"status"`
}

var (
	menu = []MenuItem{
		{ID: 1, Name: "Nasi Goreng", Price: 25000},
		{ID: 2, Name: "Mie Goreng", Price: 20000},
		{ID: 3, Name: "Ayam Bakar", Price: 30000},
	}
	orders    = make(map[int]Order)
	orderID   = 1
	orderLock sync.Mutex
)

func main() {
	http.HandleFunc("/menu", getMenu)
	http.HandleFunc("/order", createOrder)
	http.HandleFunc("/order/add", addItemToOrder)
	http.HandleFunc("/order/pay", payOrder)
	http.HandleFunc("/order/history", getOrderHistory)
	http.HandleFunc("/order/status", updateOrderStatus)

	fmt.Println("Server berjalan di port 9090")
	http.ListenAndServe(":9090", nil)
}

// Menampilkan menu makanan
func getMenu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menu)
}

// Membuat pesanan baru
func createOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var items []int
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderLock.Lock()
	defer orderLock.Unlock()

	var orderItems []MenuItem
	var total float64

	for _, itemID := range items {
		for _, menuItem := range menu {
			if menuItem.ID == itemID {
				orderItems = append(orderItems, menuItem)
				total += menuItem.Price
			}
		}
	}

	newOrder := Order{
		ID:     orderID,
		Items:  orderItems,
		Total:  total,
		Status: StatusProses,
	}
	orders[orderID] = newOrder
	orderID++

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newOrder)
}

// Menambahkan item ke pesanan
func addItemToOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		OrderID int   `json:"order_id"`
		Items   []int `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderLock.Lock()
	defer orderLock.Unlock()

	order, exists := orders[data.OrderID]
	if !exists {
		http.Error(w, "Pesanan tidak ditemukan", http.StatusNotFound)
		return
	}

	for _, itemID := range data.Items {
		for _, menuItem := range menu {
			if menuItem.ID == itemID {
				order.Items = append(order.Items, menuItem)
				order.Total += menuItem.Price
			}
		}
	}

	orders[data.OrderID] = order

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// Melakukan pembayaran pesanan
func payOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		OrderID int `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderLock.Lock()
	defer orderLock.Unlock()

	order, exists := orders[data.OrderID]
	if !exists {
		http.Error(w, "Pesanan tidak ditemukan", http.StatusNotFound)
		return
	}

	if order.Status != StatusProses {
		http.Error(w, "Pesanan sudah dibayar atau sedang diproses", http.StatusBadRequest)
		return
	}

	order.Status = StatusAntar
	orders[data.OrderID] = order

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Pembayaran berhasil. Pesanan sedang diantar.",
	})
}

// Melihat riwayat pesanan
func getOrderHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orderLock.Lock()
	defer orderLock.Unlock()

	var allOrders []Order
	for _, order := range orders {
		allOrders = append(allOrders, order)
	}

	json.NewEncoder(w).Encode(allOrders)
}

// Mengedit status pesanan
func updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		OrderID int         `json:"order_id"`
		Status  OrderStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderLock.Lock()
	defer orderLock.Unlock()

	order, exists := orders[data.OrderID]
	if !exists {
		http.Error(w, "Pesanan tidak ditemukan", http.StatusNotFound)
		return
	}

	order.Status = data.Status
	orders[data.OrderID] = order

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
