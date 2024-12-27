package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

type TestCase struct {
	Receipt        Receipt `json:"receipt"`
	ExpectedPoints int     `json:"expectedPoints"`
}

func TestReceiptProcessor(t *testing.T) {
	// Read test cases from JSON file
	testCasesJSON, err := os.ReadFile("test_cases.json")
	if err != nil {
		t.Fatalf("Failed to read test cases: %v", err)
	}

	var testCases []TestCase
	err = json.Unmarshal(testCasesJSON, &testCases)
	if err != nil {
		t.Fatalf("Failed to unmarshal test cases: %v", err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("TestCase_%d", i+1), func(t *testing.T) {
			// Process receipt
			receiptID := processReceiptAndGetID(t, tc.Receipt)

			// Get points
			points := getPointsForReceipt(t, receiptID)

			// Compare points
			if points != tc.ExpectedPoints {
				t.Errorf("Expected %d points, but got %d", tc.ExpectedPoints, points)
			}
		})
	}
}

func processReceiptAndGetID(t *testing.T, receipt Receipt) string {
	payload, _ := json.Marshal(receipt)
	req, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/receipts/process", processReceipt).Methods("POST")
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response ReceiptResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID == "" {
		t.Fatalf("Expected non-empty ID in response")
	}

	return response.ID
}

func getPointsForReceipt(t *testing.T, id string) int {
	req, _ := http.NewRequest("GET", "/receipts/"+id+"/points", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/receipts/{id}/points", getPoints).Methods("GET")
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response PointsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	return response.Points
}
