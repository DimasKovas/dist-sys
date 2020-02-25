package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

type Item struct {
	Title    string `json:"title"`
	ID       uint64 `json:"id"`
	Category string `json:"category"`
}

type ItemList struct {
	Items []Item `json:"items"`
}

var GlobalItems ItemList
var GlobalGuard sync.RWMutex

func IsStringIn(val string, arr []string) bool {
	for _, item := range arr {
		if val == item {
			return true
		}
	}
	return false
}

func IsUint64In(val uint64, arr []uint64) bool {
	for _, item := range arr {
		if val == item {
			return true
		}
	}
	return false
}

func (list ItemList) FilterByTitles(titles []string) ItemList {
	var result ItemList
	result.Items = make([]Item, 0)
	for _, item := range list.Items {
		if IsStringIn(item.Title, titles) {
			result.Items = append(result.Items, item)
		}
	}
	return result
}

func (list ItemList) FilterByCategories(categories []string) ItemList {
	var result ItemList
	result.Items = make([]Item, 0)
	for _, item := range list.Items {
		if IsStringIn(item.Category, categories) {
			result.Items = append(result.Items, item)
		}
	}
	return result
}

func (list ItemList) FilterByIds(ids []uint64) ItemList {
	var result ItemList
	result.Items = make([]Item, 0)
	for _, item := range list.Items {
		if IsUint64In(item.ID, ids) {
			result.Items = append(result.Items, item)
		}
	}
	return result
}

func ConvertIdsFromStrings(ids []string) ([]uint64, error) {
	var result []uint64
	for _, strid := range ids {
		id, err := strconv.ParseUint(strid, 10, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

func ItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		GlobalGuard.RLock()
		defer GlobalGuard.RUnlock()
		param := r.URL.Query()
		values, ok := param["id"]
		if !ok || len(values) != 1 {
			http.Error(w, "You should specify exactly one 'id' parameter", http.StatusBadRequest)
			return
		}
		ids, err := ConvertIdsFromStrings(values)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result := GlobalItems.FilterByIds(ids)
		if len(result.Items) != 1 {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result.Items[0])
	} else {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
	}
}

func ItemListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		GlobalGuard.RLock()
		defer GlobalGuard.RUnlock()
		result := GlobalItems
		param := r.URL.Query()
		values, ok := param["title"]
		if ok {
			result = result.FilterByTitles(values)
		}
		values, ok = param["category"]
		if ok {
			result = result.FilterByCategories(values)
		}
		values, ok = param["id"]
		if ok {
			ids, err := ConvertIdsFromStrings(values)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result = result.FilterByIds(ids)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	} else if r.Method == http.MethodPost {
		GlobalGuard.Lock()
		defer GlobalGuard.Unlock()
		var item Item
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var ids [1]uint64
		ids[0] = item.ID
		if len(GlobalItems.FilterByIds(ids[:]).Items) != 0 {
			http.Error(w, "Item with this ID already exists", http.StatusBadRequest)
			return
		}
		GlobalItems.Items = append(GlobalItems.Items, item)
	} else {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
	}
}

func main() {
	GlobalItems.Items = make([]Item, 0)
	http.HandleFunc("/items", ItemListHandler)
	http.HandleFunc("/items/", ItemHandler)
	http.ListenAndServe(":8080", nil)
}
