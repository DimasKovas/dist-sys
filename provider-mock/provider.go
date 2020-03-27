package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type okResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
}

func respondOK(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type errorResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
}

func respondWithError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errorResponse{"Error", statusCode, err.Error()})
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())
	params := r.URL.Query()
	if params["msg"] == nil || len(params["msg"]) != 1 {
		respondWithError(w, errors.New("Message is not specified"), 203)
		return
	}
	msg := params["msg"][0]
	if params["to"] == nil || len(params["to"]) == 0 {
		respondWithError(w, errors.New("Recipient is not specified"), 202)
		return
	}
	for _, to := range params["to"] {
		log.Printf("To: %s, Message: '%s'", to, msg)
	}
	respondOK(w, okResponse{"OK", 100})
}

func main() {
	http.HandleFunc("/send", sendHandler)
	log.Println("Provider-mock started")
	log.Panic(http.ListenAndServe(":8080", nil))
}
