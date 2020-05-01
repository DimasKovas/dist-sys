package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"authentication/dbclient"
	"authentication/mqclient"

	"golang.org/x/crypto/bcrypt"
)

var db dbclient.Client
var mq mqclient.Client

type Config struct {
	TokenLength          uint
	RefreshTokenLifeTime time.Duration
	AccessTokenLifeTime  time.Duration
	ConfirmTokenLifeTime time.Duration
	ConfirmAddress       string
}

var conf Config

func (c *Config) Load() error {
	tlen, err := strconv.ParseUint(os.Getenv("TOKEN_LENGTH"), 10, 64)
	c.TokenLength = uint(tlen)
	c.RefreshTokenLifeTime, err = time.ParseDuration(os.Getenv("REFRESH_TOKEN_LIFE_TIME"))
	if err != nil {
		return err
	}
	c.AccessTokenLifeTime, err = time.ParseDuration(os.Getenv("ACCESS_TOKEN_LIFE_TIME"))
	if err != nil {
		return err
	}
	c.ConfirmTokenLifeTime, err = time.ParseDuration(os.Getenv("CONFIRM_TOKEN_LIFE_TIME"))
	if err != nil {
		return err
	}
	c.ConfirmAddress = os.Getenv("CONFIRM_ADDRESS")
	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(lenght uint) string {
	b := make([]byte, (lenght+1)/2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
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

func checkPermission(w http.ResponseWriter, r *http.Request, perm string) bool {
	token := r.Header.Get("auth")
	perms, err := doValidate(token)
	if err != nil {
		respondWithError(w, err, http.StatusUnauthorized)
		return false
	}
	for _, p := range perms.Permissions {
		if p == perm {
			return true
		}
	}
	respondWithError(w, errors.New("Not enough permissions"), http.StatusForbidden)
	return false
}

func sendConfirmationMessage(username string, phone string) error {
	token := generateToken(conf.TokenLength)
	tinfo := dbclient.TokenInfo{
		token,
		time.Now().Add(conf.ConfirmTokenLifeTime),
		dbclient.CONFIRM,
		username,
	}
	err := db.AddNewToken(tinfo)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("To confirm registration folow the link: %s/%s", conf.ConfirmAddress, token)
	err = mq.SendMessage(mqclient.Message{phone, msg})
	if err != nil {
		return err
	}
	return nil
}

type postSignUpRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
}

type postSignUpResponse struct {
	Message string `json:"message"`
}

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
	user.PhoneNumber = request.PhoneNumber
	user.PhoneConfirmed = false
	user.Permissions = append(user.Permissions, "read")
	err = db.AddUser(user)
	if err != nil {
		switch err {
		case dbclient.ErrUserAlreadyExists:
			respondWithError(w, err, http.StatusBadRequest)
		default:
			respondWithError(w, err, http.StatusInternalServerError)
		}
		return
	}
	err = sendConfirmationMessage(user.Username, user.PhoneNumber)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	respondOK(w, postSignUpResponse{"We sent the confirmation link to your phone."})
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

var ErrNotValid = errors.New("Username or password is not valid")
var ErrPhoneNotConfirmed = errors.New("Your phone number is not confirmed. Follow the link in the sms")

func postSignInHandler(w http.ResponseWriter, r *http.Request) {
	var request postSignInRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	user, err := db.GetUser(request.Username)
	if err != nil {
		switch err {
		case dbclient.ErrNotFound:
			respondWithError(w, ErrNotValid, http.StatusUnauthorized)
		default:
			respondWithError(w, err, http.StatusInternalServerError)
		}
		return
	}
	if !checkPasswordHash(request.Password, user.PassHash) {
		respondWithError(w, ErrNotValid, http.StatusUnauthorized)
		return
	}
	if !user.PhoneConfirmed {
		err = sendConfirmationMessage(user.Username, user.PhoneNumber)
		if err != nil {
			respondWithError(w, err, http.StatusInternalServerError)
			return
		}
		respondWithError(w, ErrPhoneNotConfirmed, http.StatusForbidden)
		return
	}
	refresh := generateToken(conf.TokenLength)
	access := generateToken(conf.TokenLength)
	rtinfo := dbclient.TokenInfo{
		refresh,
		time.Now().Add(conf.RefreshTokenLifeTime),
		dbclient.REFRESH,
		request.Username,
	}
	atinfo := dbclient.TokenInfo{
		access,
		time.Now().Add(conf.AccessTokenLifeTime),
		dbclient.ACCESS,
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

type getValidateRequest struct{}

type getValidateResponse struct {
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
}

func doValidate(token string) (getValidateResponse, error) {
	tinfo, err := db.GetTokenInfo(token)
	if err != nil {
		switch err {
		case dbclient.ErrNotFound:
			return getValidateResponse{}, errors.New("Token is not valid")
		default:
			return getValidateResponse{}, err
		}
	}
	if tinfo.Type != dbclient.ACCESS {
		return getValidateResponse{}, errors.New("Should provide access token")
	}
	if tinfo.ExpTime.Before(time.Now()) {
		return getValidateResponse{}, errors.New("Token has expired")
	}
	uinfo, err := db.GetUser(tinfo.Username)
	if err != nil {
		return getValidateResponse{}, err
	}
	result := getValidateResponse{
		Username:    uinfo.Username,
		Permissions: uinfo.Permissions,
	}
	return result, nil
}

func getValidateHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("auth")
	result, err := doValidate(token)
	if err != nil {
		respondWithError(w, err, http.StatusUnauthorized)
		return
	}
	respondOK(w, result)
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
		switch err {
		case dbclient.ErrNotFound:
			respondWithError(w, errors.New("Token is not valid"), http.StatusForbidden)
		default:
			respondWithError(w, err, http.StatusInternalServerError)
		}
		return
	}
	if rtinfo.Type != dbclient.REFRESH {
		respondWithError(w, errors.New("Should provide refresh token"), http.StatusBadRequest)
		return
	}
	if rtinfo.ExpTime.Before(time.Now()) {
		respondWithError(w, errors.New("Token has expired"), http.StatusForbidden)
		return
	}
	access := generateToken(conf.TokenLength)
	atinfo := dbclient.TokenInfo{
		access,
		time.Now().Add(conf.AccessTokenLifeTime),
		dbclient.ACCESS,
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

type putSetPermissionsRequest struct {
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
}

type putSetPermissionsResponse struct{}

func putSetPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	if !checkPermission(w, r, "manage") {
		return
	}
	var request putSetPermissionsRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	err = db.SetPermissions(request.Username, request.Permissions)
	if err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}
	respondOK(w, putSetPermissionsResponse{})
}

func generalSetPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		putSetPermissionsHandler(w, r)
	default:
		w.Header().Add("Allow", "PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type confirmResponse struct {
	Message string `json:"message"`
}

func confirmHandler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/confirm/")
	tinfo, err := db.GetTokenInfo(token)
	if err != nil {
		switch err {
		case dbclient.ErrNotFound:
			respondWithError(w, errors.New("Token is not valid"), http.StatusUnauthorized)
		default:
			respondWithError(w, err, http.StatusInternalServerError)
		}
		return
	}
	if tinfo.Type != dbclient.CONFIRM {
		respondWithError(w, errors.New("Should provide confirm token"), http.StatusBadRequest)
		return
	}
	if tinfo.ExpTime.Before(time.Now()) {
		respondWithError(w, errors.New("Token has expired"), http.StatusUnauthorized)
		return
	}
	err = db.ConfirmPhoneNumber(tinfo.Username)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}
	respondOK(w, confirmResponse{"Registration has been successfully confirmed"})
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
	mq, err = mqclient.CreateMqClient()
	if err != nil {
		log.Panic(err)
	}
	http.HandleFunc("/signup", generalSignUpHandler)
	http.HandleFunc("/signin", generalSignInHandler)
	http.HandleFunc("/validate", generalValidateHandler)
	http.HandleFunc("/refresh", generalRefreshHandler)
	http.HandleFunc("/set_permissions", generalSetPermissionsHandler)
	http.HandleFunc("/confirm/", confirmHandler)
	log.Println("Auth-server started")
	go RunRpcServer()
	log.Panic(http.ListenAndServe(":8080", nil))
}
