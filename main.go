package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Receipt struct {
	Retailer     string  `json:"retailer"`
	PurchaseDate string  `json:"purchaseDate"`
	PurchaseTime string  `json:"purchaseTime"`
	Items        []Item  `json:"items"`
	Total        string  `json:"total"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type ReceiptResponse struct {
	ID string `json:"id"`
}

type PointsResponse struct {
	Points int `json:"points"`
}

var receipts = make(map[string]Receipt)
var points = make(map[string]int)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/receipts/process", processReceipt).Methods("POST")
	r.HandleFunc("/receipts/{id}/points", getPoints).Methods("GET")

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func processReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	receipts[id] = receipt
	points[id] = calculatePoints(receipt)

	response := ReceiptResponse{ID: id}
	json.NewEncoder(w).Encode(response)
}

func getPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if _, exists := receipts[id]; !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	response := PointsResponse{Points: points[id]}
	json.NewEncoder(w).Encode(response)
}

func calculatePoints(receipt Receipt) int {
	totalPoints := 0

	// Rule 1: One point for every alphanumeric character in the retailer name
	alphanumeric := regexp.MustCompile(`[a-zA-Z0-9]`)
	totalPoints += len(alphanumeric.FindAllString(receipt.Retailer, -1))

	// Rule 2: 50 points if the total is a round dollar amount with no cents
	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if total == float64(int(total)) {
		totalPoints += 50
	}

	// Rule 3: 25 points if the total is a multiple of 0.25
	if math.Mod(total*100, 25) == 0 {
		totalPoints += 25
	}

	// Rule 4: 5 points for every two items on the receipt
	totalPoints += (len(receipt.Items) / 2) * 5

	// Rule 5: If the trimmed length of the item description is a multiple of 3,
	// multiply the price by 0.2 and round up to the nearest integer
	for _, item := range receipt.Items {
		if len(strings.TrimSpace(item.ShortDescription))%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			totalPoints += int(math.Ceil(price * 0.2))
		}
	}

	// Rule 6: 6 points if the day in the purchase date is odd
	purchaseDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
	if purchaseDate.Day()%2 != 0 {
		totalPoints += 6
	}

	// Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm
	purchaseTime, _ := time.Parse("15:04", receipt.PurchaseTime)
	if purchaseTime.Hour() >= 14 && purchaseTime.Hour() < 16 {
		totalPoints += 10
	}

	// Rule 8: 5 points if the total is greater than 10.00
	if total > 10.00 {
		totalPoints += 5
	}

	return totalPoints
}
