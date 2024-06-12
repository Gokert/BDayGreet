package middleware

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"vk-rest/pkg/models"
	httpResponse "vk-rest/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "userId"

type IMiddleware interface {
	GetUserId(ctx context.Context, sid string) (uint64, error)
}

type Middleware struct {
	Lg   *logrus.Logger
	Core IMiddleware
}

func (m *Middleware) AuthCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("session_id")
		if errors.Is(err, http.ErrNoCookie) {
			response := models.Response{Status: http.StatusUnauthorized, Body: models.ErrorResponse{Error: "Not authorized"}}
			httpResponse.SendResponse(w, r, &response, m.Lg)
			return
		}

		userId, err := m.Core.GetUserId(r.Context(), session.Value)
		if err != nil {
			m.Lg.Error("auth check error", "err", err.Error())
			response := models.Response{Status: http.StatusUnauthorized, Body: models.ErrorResponse{Error: "Not authorized"}}
			httpResponse.SendResponse(w, r, &response, m.Lg)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), UserIDKey, userId))
		if userId == 0 {
			response := models.Response{Status: http.StatusUnauthorized, Body: nil}
			httpResponse.SendResponse(w, r, &response, m.Lg)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) MethodCheck(next http.Handler, method string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			response := models.Response{Status: http.StatusMethodNotAllowed, Body: models.ErrorResponse{Error: "Method not allowed"}}
			httpResponse.SendResponse(w, r, &response, m.Lg)
			return
		}
		next.ServeHTTP(w, r)
	})
}
