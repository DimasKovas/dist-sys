package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"common/auth"
	"common/dbclient"
)

type Item = dbclient.Item

var db dbclient.Client
var ac auth.AuthClient

func respondOK(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type errorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errorResponse{err.Error()})
}

func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func extractIndexFromUrl(url string, pref string) (uint64, error) {
	return parseUint(strings.TrimPrefix(url, pref))
}

func checkPermission(w http.ResponseWriter, r *http.Request) bool {
	token := r.Header.Get("auth")
	err := ac.Validate(token)
	if err != nil {
		if e, ok := err.(*auth.ErrResponseWithStatus); ok {
			respondWithError(w, e.RemoteError, e.StatusCode)
		} else {
			respondWithError(w, err, http.StatusInternalServerError)
		}
		return false
	}
	return true
}

type getItemResponse = Item

func getItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := extractIndexFromUrl(r.URL.Path, "/item/")
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	item, err := db.GetItem(id)
	if err != nil {
		switch err {
		case dbclient.ErrNotFound:
			respondWithError(w, err, http.StatusNotFound)
		default:
			respondWithError(w, err, http.StatusInternalServerError)
		}
		return
	}
	respondOK(w, item)
}

type postItemResponse struct {
	ID uint64 `json:"id"`
}

func postItemHandler(w http.ResponseWriter, r *http.Request) {
	if !checkPermission(w, r) {
		return
	}
	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	id, err := db.NewItem(item)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	respondOK(w, postItemResponse{id})
}

type putItemResponse struct{}

func putItemHandler(w http.ResponseWriter, r *http.Request) {
	if !checkPermission(w, r) {
		return
	}
	id, err := extractIndexFromUrl(r.URL.Path, "/item/")
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	var item Item
	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	item.ID = id
	err = db.UpdateItem(item)
	if err != nil {
		switch err {
		case dbclient.ErrNotFound:
			respondWithError(w, err, http.StatusNotFound)
		default:
			respondWithError(w, err, http.StatusInternalServerError)
		}
		return
	}
	respondOK(w, putItemResponse{})
}

type deleteItemResponse struct{}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	if !checkPermission(w, r) {
		return
	}
	id, err := extractIndexFromUrl(r.URL.Path, "/item/")
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	err = db.DeleteItem(id)
	if err != nil {
		switch err {
		case dbclient.ErrNotFound:
			respondWithError(w, err, http.StatusNotFound)
		default:
			respondWithError(w, err, http.StatusInternalServerError)
		}
		return
	}
	respondOK(w, deleteItemResponse{})
}

func generalItemHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getItemHandler(w, r)
	case "POST":
		postItemHandler(w, r)
	case "PUT":
		putItemHandler(w, r)
	case "DELETE":
		deleteItemHandler(w, r)
	default:
		w.Header().Add("Allow", "GET, POST, PUT, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func extractUintFromParams(params url.Values, name string) (uint64, error) {
	if len(params[name]) != 1 {
		return 0, fmt.Errorf("Only one parameter '%v' can be specified", name)
	}
	return parseUint(params[name][0])
}

type getItemsResponse struct {
	Count int    `json:"count"`
	Items []Item `json:"items"`
}

func getItemsHandler(w http.ResponseWriter, r *http.Request) {
	var options dbclient.GetItemListOptions
	params := r.URL.Query()
	if params["offset"] != nil {
		offset, err := extractUintFromParams(params, "offset")
		if err != nil {
			respondWithError(w, err, http.StatusBadRequest)
			return
		}
		options.Offset = offset
	}
	limit := uint64(1000)
	if params["limit"] != nil {
		sLimit, err := extractUintFromParams(params, "limit")
		if err != nil {
			respondWithError(w, err, http.StatusBadRequest)
			return
		}
		if sLimit < limit {
			limit = sLimit
		}
	}
	options.Limit = limit
	if params["category"] != nil {
		options.Categories = params["category"]
	}
	items, err := db.GetItemList(options)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	size, err := db.GetItemListSize()
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	respondOK(w, getItemsResponse{size, items})
}

func generalItemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getItemsHandler(w, r)
	default:
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	var err error
	ac, err = auth.CreateAuthClient()
	if err != nil {
		log.Panic(err)
	}
	db, err = dbclient.CreateDbClient()
	if err != nil {
		log.Panic(err)
	}
	http.HandleFunc("/items", generalItemsHandler)
	http.HandleFunc("/item", generalItemHandler)
	http.HandleFunc("/item/", generalItemHandler)
	log.Println("Item-storage started")
	log.Panic(http.ListenAndServe(":8080", nil))
}
