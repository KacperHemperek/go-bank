package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type APIServer struct {
	addr     string
	store    Storage
	validate *validator.Validate
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	log.Println("Server is listening on: ", s.addr)

	router.HandleFunc("/account", HandleFunc(s.handleAccount))

	log.Fatal(http.ListenAndServe(s.addr, router))
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleGetAccount(w, r)
	}
	if r.Method == http.MethodPost {
		return s.handleCreateAccount(w, r)
	}
	if r.Method == http.MethodDelete {
		return s.handleDeleteAccount(w, r)
	}

	return ApiError{
		Err:    "Method not allowed",
		Status: http.StatusMethodNotAllowed,
	}
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	account := NewAccount("Kacper", "Hemp")

	return WriteJson(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	accountReq := &AccountCreateRequest{}
	if err := DecodeJson(r, accountReq); err != nil {
		return ApiError{Err: "Could not decode request", Status: http.StatusBadRequest, Cause: err}
	}
	if err := s.validate.Struct(accountReq); err != nil {
		return ApiError{Err: "Invalid request", Status: http.StatusBadRequest, Cause: err}
	}

	acc := NewAccount(accountReq.FirstName, accountReq.LastName)

	result, err := s.store.CreateAccount(acc)

	if err != nil {
		return ApiError{Err: "Could not create account", Status: http.StatusInternalServerError, Cause: err}
	}

	return WriteJson(w, http.StatusCreated, result)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func DecodeJson(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
	Err    string `json:"error"`
	Status int    `json:"status"`
	Cause  error  `json:"-"`
}

func (e ApiError) Error() string {
	return fmt.Sprintf("%s (%d)", e.Err, e.Status)
}

func NewAPIServer(addr string, store Storage) *APIServer {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &APIServer{
		addr:     addr,
		store:    store,
		validate: validate,
	}
}

func HandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		if err := f(w, r); err != nil {
			fmt.Printf("Error %s [%s]: %s %s\n", r.Method, r.URL, err.Error(), time.Since(now))
			if apiErr, ok := err.(ApiError); ok {
				fmt.Println("cause: ", apiErr.Cause)
				WriteJson(w, apiErr.Status, apiErr)
				return
			}

			WriteJson(w, http.StatusInternalServerError, ApiError{Err: err.Error(), Status: http.StatusBadRequest})
			return
		}
		fmt.Printf("Success [%s]: %s %s\n", r.Method, r.URL, time.Since(now))
	}
}
