package models

type SigninRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type SignupRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Birthday string `json:"birthday"`
}

type SubRequest struct {
	UserFromId uint64 `json:"user_from_id"`
	UserToId   uint64 `json:"user_to_id"`
}

type UnSubRequest struct {
	UserFromId uint64 `json:"user_from_id"`
	UserToId   uint64 `json:"user_to_id"`
}
