package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/DimasKovas/dist-sys/dbclient"
)

type Item dbclient.Item

var db dbclient.Client

type ItemList struct {
	Items []Item `json:"items"`
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
	stdid := r.URL.Path[len("/item/"):]
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

func responseOK(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func responseError(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(err)
}

func extractIndexFromUrl(url string, pref string) (uint64, error) {
	return strconv.ParseUint(strings.TrimPrefix(url, pref), 10, 64)
}

func getItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := extractIndexFromUrl(r.URL.Path, "/item/")
	if err != nil {
		responseError(w, err, http.StatusBadRequest)
		return
	}
	item, err := db.GetItem(id)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	responseOK(w, item)
}

func postItemHandler(w http.ResponseWriter, r *http.Request) {
	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		responseError(w, err, http.StatusBadRequest)
		return
	}
	id, err := db.NewItem(item)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	responseOK(w, id)
}

func putItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := extractIndexFromUrl(r.URL.Path, "/item/")
	if err != nil {
		responseError(w, err, http.StatusBadRequest)
		return
	}
	var item Item
	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		responseError(w, err, http.StatusBadRequest)
		return
	}
	item.ID = id
	err = db.UpdateItem(item)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	responseOK(w, nil)
}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := extractIndexFromUrl(r.URL.Path, "/item/")
	if err != nil {
		responseError(w, err, http.StatusBadRequest)
		return
	}
	err = db.DeleteItem(id)
	if err != nil {
		responseError(w, err, http.StatusBadRequest)
		return
	}
	responseOK(w, nil)
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

func getItemsHandler(w http.ResponseWriter, r *http.Request) {

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
	GlobalItems.Items = make([]Item, 0)
	http.HandleFunc("/items", ItemListHandler)
	http.HandleFunc("/item", generalItemHandler)
	http.ListenAndServe(":8080", nil)
}
