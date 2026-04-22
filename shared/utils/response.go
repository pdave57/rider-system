package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func OK(w http.ResponseWriter, msg string, data interface{}) {
	WriteJSON(w, http.StatusOK, Response{Success: true, Message: msg, Data: data})
}

func Created(w http.ResponseWriter, msg string, data interface{}) {
	WriteJSON(w, http.StatusCreated, Response{Success: true, Message: msg, Data: data})
}

func BadRequest(w http.ResponseWriter, msg string) {
	WriteJSON(w, http.StatusBadRequest, Response{Success: false, Error: msg})
}

func Unauthorized(w http.ResponseWriter, msg string) {
	WriteJSON(w, http.StatusUnauthorized, Response{Success: false, Error: msg})
}

func Forbidden(w http.ResponseWriter, msg string) {
	WriteJSON(w, http.StatusForbidden, Response{Success: false, Error: msg})
}

func NotFound(w http.ResponseWriter, msg string) {
	WriteJSON(w, http.StatusNotFound, Response{Success: false, Error: msg})
}

func InternalError(w http.ResponseWriter, msg string) {
	WriteJSON(w, http.StatusInternalServerError, Response{Success: false, Error: msg})
}

func Conflict(w http.ResponseWriter, msg string) {
	WriteJSON(w, http.StatusConflict, Response{Success: false, Error: msg})
}

func DecodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
