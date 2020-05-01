package main

import (
	"common/auth"
	"common/dbclient"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"item-uploader/mqclient"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Config struct {
	BatchSize int
}

func (c *Config) Load() error {
	batch_size, err := strconv.ParseInt(os.Getenv("BATCH_SIZE"), 10, 32)
	if err != nil {
		return err
	}
	c.BatchSize = int(batch_size)
	return nil
}

var conf Config
var mq mqclient.Client
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

func checkPermission(w http.ResponseWriter, r *http.Request, perm string) bool {
	token := r.Header.Get("auth")
	err := ac.CheckPermission(token, perm)
	if err != nil {
		respondWithError(w, err, http.StatusForbidden)
		return false
	}
	return true
}

type postImportResponse struct{}

func postImportHandler(w http.ResponseWriter, r *http.Request) {
	if !checkPermission(w, r, "write") {
		return
	}
	reader, err := r.MultipartReader()
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	part, err := reader.NextPart()
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	if part.FormName() != "file" {
		log.Println("Expected file, got " + part.FormName())
		respondWithError(w, errors.New("File is expected"), http.StatusBadRequest)
		return
	}
	csvreader := csv.NewReader(part)
	names, err := csvreader.Read()
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	name_to_index := make(map[string]int)
	for i, name := range names {
		name_to_index[name] = i
	}
	item_fields := [2]string{"title", "category"}
	for _, field := range item_fields {
		if _, ok := name_to_index[field]; !ok {
			respondWithError(w, errors.New("Item should contain '"+field+"' field"), http.StatusBadRequest)
			return
		}
	}
	closed := false

	var batch []dbclient.Item

	for !closed {
		values, err := csvreader.Read()
		if err == io.EOF {
			closed = true
		} else if err != nil {
			respondWithError(w, err, http.StatusBadRequest)
			return
		} else {
			var item dbclient.Item
			item.Title = values[name_to_index["title"]]
			item.Category = values[name_to_index["category"]]
			if index, ok := name_to_index["universal_code"]; ok {
				item.UniversalCode = values[index]
			}
			batch = append(batch, item)
		}
		if len(batch) == conf.BatchSize || (closed && len(batch) != 0) {
			mq.SendImportBatch(batch)
		}
	}
	respondOK(w, postImportResponse{})
}

func generalImportHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		postImportHandler(w, r)
	default:
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	var err error
	err = conf.Load()
	if err != nil {
		log.Panic(err)
	}
	ac, err = auth.CreateAuthClient()
	if err != nil {
		log.Panic(err)
	}
	mq, err = mqclient.CreateMqClient()
	if err != nil {
		log.Panic(err)
	}
	http.HandleFunc("/import", generalImportHandler)
	log.Println("Item-uploader started")
	log.Panic(http.ListenAndServe(":8080", nil))
}
