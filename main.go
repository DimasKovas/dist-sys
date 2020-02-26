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

type FilterFunc func(Item) bool

type ItemList struct {
	Items []Item `json:"items"`
}

var GlobalItems ItemList
var GlobalGuard sync.RWMutex

func (list ItemList) FilterBy(filter FilterFunc) ItemList {
	var result ItemList
	result.Items = make([]Item, 0)
	for _, item := range list.Items {
		if filter(item) {
			result.Items = append(result.Items, item)
		}
	}
	return result
}

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
	return list.FilterBy(func(item Item) bool {
		return IsStringIn(item.Title, titles)
	})
}

func (list ItemList) FilterByCategories(categories []string) ItemList {
	return list.FilterBy(func(item Item) bool {
		return IsStringIn(item.Category, categories)
	})
}

func (list ItemList) FilterByIds(ids []uint64) ItemList {
	return list.FilterBy(func(item Item) bool {
		return IsUint64In(item.ID, ids)
	})
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
	stdid := r.URL.Path[len("/items/"):]
	var ids [1]uint64
	id, err := strconv.ParseUint(stdid, 10, 64)
	ids[0] = id
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet {
		// prefix is /items/, skip it

		GlobalGuard.RLock()
		defer GlobalGuard.RUnlock()
		result := GlobalItems.FilterByIds(ids[:])
		if len(result.Items) != 1 {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result.Items[0])
	} else if r.Method == http.MethodDelete {
		GlobalGuard.Lock()
		defer GlobalGuard.Unlock()
		oldLen := len(GlobalItems.Items)
		GlobalItems = GlobalItems.FilterBy(func(item Item) bool {
			return item.ID != id
		})
		if len(GlobalItems.Items) == oldLen {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
	} else if r.Method == http.MethodPut {
		var item Item
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if id != item.ID {
			http.Error(w, "ID dismiss", http.StatusBadRequest)
			return
		}
		GlobalGuard.Lock()
		defer GlobalGuard.Unlock()
		oldLen := len(GlobalItems.Items)
		GlobalItems = GlobalItems.FilterBy(func(item Item) bool {
			return item.ID != id
		})
		if len(GlobalItems.Items) == oldLen {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		GlobalItems.Items = append(GlobalItems.Items, item)
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
