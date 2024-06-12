package models

type Response struct {
	Status int `json:"status"`
	Body   any `json:"body"`
}

type AuthCheckResponse struct {
	Login string `json:"login"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
