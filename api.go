package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
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

	router.HandleFunc("/accounts", HandleFunc(s.handleAccounts))

	router.HandleFunc("/accounts/{id}", HandleFunc(s.HandleAccount))

	router.HandleFunc("/transfer", HandleFunc(s.handleTransfer))

	log.Fatal(http.ListenAndServe(s.addr, router))
}

func (s *APIServer) handleAccounts(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleGetAccounts(w, r)
	}
	if r.Method == http.MethodPost {
		return s.handleCreateAccount(w, r)
	}

	return ApiError{
		Err:    "Method not allowed",
		Status: http.StatusMethodNotAllowed,
	}
}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, _ *http.Request) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return ApiError{Err: "Could not get accounts", Status: http.StatusInternalServerError, Cause: err}
	}

	response := ResObj{
		"accounts": accounts,
	}

	return WriteJson(w, http.StatusOK, response)
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

func (s *APIServer) HandleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleGetAccount(w, r)
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
	id, err := getId(r)

	if err != nil {
		return err
	}

	acc, err := s.store.GetAccountByID(id)

	if err != nil {
		return ApiError{Err: "Account not found", Status: http.StatusNotFound, Cause: err}
	}

	return WriteJson(w, http.StatusOK, acc)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)

	if err != nil {
		return err
	}

	err = s.store.DeleteAccount(id)
	if err != nil {
		return ApiError{Err: "Could not delete account", Status: http.StatusInternalServerError, Cause: err}
	}

	return WriteJson(w, http.StatusOK, ResObj{"status": "deleted"})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {

	if r.Method != http.MethodPost {
		return ApiError{Err: "Method not allowed", Status: http.StatusMethodNotAllowed}
	}

	transferReq := &TransferRequest{}
	defer r.Body.Close()

	if err := DecodeJson(r, transferReq); err != nil {
		return ApiError{Err: "Could not decode request", Status: http.StatusBadRequest, Cause: err}
	}

	if err := s.validate.Struct(transferReq); err != nil {
		return ApiError{Err: "Invalid request", Status: http.StatusBadRequest, Cause: err}
	}

	return WriteJson(w, http.StatusOK, ResObj{"status": "transferred successfully"})
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

func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, ApiError{Err: "ID must be a number", Status: http.StatusBadRequest, Cause: err}
	}
	return id, nil
}

type ResObj map[string]interface{}
