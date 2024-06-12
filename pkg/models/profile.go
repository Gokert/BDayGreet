package models

type UserItem struct {
	Id       uint64 `json:"id"`
	Login    string `json:"login"`
	Email    string `json:"email"`
	Birthday string `json:"birthday"`
}
