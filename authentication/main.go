package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"authentication/dbclient"

	"golang.org/x/crypto/bcrypt"
)

var db dbclient.Client

type Config struct {
	RefreshTokenLifeTime time.Duration
	AccessTokenLifeTime  time.Duration
}

var conf Config

func (c *Config) Load() error {
	var err error
	c.RefreshTokenLifeTime, err = time.ParseDuration(os.Getenv("REFRESH_TOKEN_LIFE_TIME"))
	if err != nil {
		return err
	}
	c.AccessTokenLifeTime, err = time.ParseDuration(os.Getenv("ACCESS_TOKEN_LIFE_TIME"))
	return err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(lenght int) string {
	return "token"
}

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
	json.NewEncoder(w).Encode(errorResponse{fmt.Sprint(err)})
}

type postSignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type postSignUpResponse struct{}

func postSignUpHandler(w http.ResponseWriter, r *http.Request) {
	var request postSignUpRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	hash, err := hashPassword(request.Password)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	var user dbclient.User
	user.Username = request.Username
	user.PassHash = hash
	user.Email = request.Email
	err = db.AddUser(user)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	respondOK(w, postSignUpRequest{})
}

func generalSignUpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		postSignUpHandler(w, r)
	default:
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type postSignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type postSignInResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

func postSignInHandler(w http.ResponseWriter, r *http.Request) {
	var request postSignInRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	user, err := db.GetUser(request.Username)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	if !checkPasswordHash(request.Password, user.PassHash) {
		respondWithError(w, errors.New("oops"), http.StatusNotFound)
		return
	}
	refresh := generateToken(10)
	access := generateToken(10)
	rtinfo := dbclient.TokenInfo{
		refresh,
		time.Now().Add(conf.RefreshTokenLifeTime),
		false,
		request.Username,
	}
	atinfo := dbclient.TokenInfo{
		refresh,
		time.Now().Add(conf.AccessTokenLifeTime),
		true,
		request.Username,
	}
	err = db.AddNewToken(rtinfo)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	err = db.AddNewToken(atinfo)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	respondOK(w, postSignInResponse{refresh, access})
}

func generalSignInHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		postSignInHandler(w, r)
	default:
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type getValidateRequest struct {
	Token string `json:"token"`
}

type getValidateResponse struct{}

func getValidateHandler(w http.ResponseWriter, r *http.Request) {
	var request getValidateRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	tinfo, err := db.GetTokenInfo(request.Token)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	if tinfo.Refresh {
		respondWithError(w, errors.New("Should provide access token"), http.StatusForbidden)
		return
	}
	if tinfo.ExpTime.Before(time.Now()) {
		respondWithError(w, errors.New("Token has expired"), http.StatusForbidden)
		return
	}
	respondOK(w, getValidateResponse{})
}

func generalValidateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getValidateHandler(w, r)
	default:
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type putRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type putRefreshResponse struct {
	AccessToken string `json:"access_token"`
}

func putRefreshHandler(w http.ResponseWriter, r *http.Request) {
	var request putRefreshRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	rtinfo, err := db.GetTokenInfo(request.RefreshToken)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	if !rtinfo.Refresh {
		respondWithError(w, errors.New("Should provide refresh token"), http.StatusBadRequest)
		return
	}
	if rtinfo.ExpTime.Before(time.Now()) {
		respondWithError(w, errors.New("Token has expired"), http.StatusForbidden)
		return
	}
	access := generateToken(10)
	atinfo := dbclient.TokenInfo{
		access,
		time.Now().Add(conf.AccessTokenLifeTime),
		false,
		rtinfo.Username,
	}
	err = db.AddNewToken(atinfo)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	respondOK(w, putRefreshResponse{access})
}

func generalRefreshHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		putRefreshHandler(w, r)
	default:
		w.Header().Add("Allow", "PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	var err error
	err = conf.Load()
	if err != nil {
		log.Panic(err)
	}
	db, err = dbclient.CreateDbClient()
	if err != nil {
		log.Panic(err)
	}
	http.HandleFunc("/signup", generalSignUpHandler)
	http.HandleFunc("/signin", generalSignInHandler)
	http.HandleFunc("/validate", generalValidateHandler)
	http.HandleFunc("/refresh", generalRefreshHandler)
	log.Println("Auth-server started")
	log.Panic(http.ListenAndServe(":8080", nil))
}
