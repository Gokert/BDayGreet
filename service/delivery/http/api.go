package delivery

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"time"
	errs "vk-rest/pkg/errors"
	"vk-rest/pkg/middleware"
	"vk-rest/pkg/models"
	httpResponse "vk-rest/pkg/response"
	"vk-rest/service/usecase/core"
)

//go:generate mockgen -source=api.go -destination=.mocks/http_api_mock.go -package=mocks
type IApi interface {
	Signin(w http.ResponseWriter, r *http.Request)
	Signup(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	AuthAccept(w http.ResponseWriter, r *http.Request)
	GetEmployees(w http.ResponseWriter, r *http.Request)
	BirthdaySub(w http.ResponseWriter, r *http.Request)
	BirthdayUnSub(w http.ResponseWriter, r *http.Request)
}

type Api struct {
	log     *logrus.Logger
	mx      *http.ServeMux
	profile usecase.IProfileCore
	session usecase.ISessionCore
	sub     usecase.ISubCore
}

func GetApi(core *usecase.Core, log *logrus.Logger) *Api {
	api := &Api{
		profile: core,
		session: core,
		sub:     core,
		log:     log,
		mx:      http.NewServeMux(),
	}

	md := &middleware.Middleware{
		Lg:   log,
		Core: core,
	}

	api.mx.Handle("/signin", md.MethodCheck(http.HandlerFunc(api.Signin), http.MethodPost))
	api.mx.Handle("/signup", md.MethodCheck(http.HandlerFunc(api.Signup), http.MethodPost))
	api.mx.Handle("/logout", md.MethodCheck(http.HandlerFunc(api.Logout), http.MethodDelete))
	api.mx.Handle("/authcheck", md.MethodCheck(http.HandlerFunc(api.AuthAccept), http.MethodGet))
	api.mx.Handle("/api/v1/employees", md.AuthCheck(md.MethodCheck(http.HandlerFunc(api.GetEmployees), http.MethodGet)))
	api.mx.Handle("/api/v1/birthday/subscribe", md.AuthCheck(md.MethodCheck(http.HandlerFunc(api.BirthdaySub), http.MethodPost)))
	api.mx.Handle("/api/v1/birthday/unsubscribe", md.AuthCheck(md.MethodCheck(http.HandlerFunc(api.BirthdayUnSub), http.MethodDelete)))

	return api
}

func (a *Api) ListenAndServe(port string) error {
	err := http.ListenAndServe(":"+port, a.mx)
	if err != nil {
		a.log.Error("ListenAndServer error: ", err.Error())
		return err
	}

	return nil
}

func (a *Api) Signin(w http.ResponseWriter, r *http.Request) {
	var request models.SigninRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.log.Error("Signin error: ", err.Error())
		response := models.Response{Status: http.StatusBadRequest, Body: models.ErrorResponse{Error: errs.ErrBadRequest}}
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		a.log.Error("Signin error: ", err.Error())
		response := models.Response{Status: http.StatusBadRequest, Body: models.ErrorResponse{Error: errs.ErrBadRequest}}
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	_, found, err := a.profile.FindUserAccount(r.Context(), request.Login, request.Password)
	if err != nil {
		a.log.Error("Signin error: ", err.Error())
		response := models.Response{Status: http.StatusInternalServerError, Body: models.ErrorResponse{Error: errs.ErrInternalServer}}
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	if !found {
		response := models.Response{Status: http.StatusUnauthorized, Body: models.ErrorResponse{Error: errs.ErrNotFoundString}}
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	session, _ := a.session.CreateSession(r.Context(), request.Login)
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    session.SID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusOK, Body: nil}, a.log)
}

func (a *Api) Signup(w http.ResponseWriter, r *http.Request) {
	var request models.SignupRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.log.Error("Signup error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusBadRequest, Body: "Bad request"}, a.log)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		a.log.Error("Signup error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusBadRequest, Body: "Bad request"}, a.log)
		return
	}

	if request.Birthday == "" || request.Email == "" || request.Password == "" || request.Login == "" {
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusBadRequest, Body: "Not have birthday, email, password or login"}, a.log)
		return
	}

	found, err := a.profile.FindUserByLogin(r.Context(), request.Login)
	if err != nil {
		a.log.Error("Signup error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusInternalServerError, Body: "Internal server error"}, a.log)
		return
	}

	if found {
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusConflict, Body: "Already exist"}, a.log)
		return
	}

	err = a.profile.CreateUserAccount(r.Context(), &request)
	if err != nil {
		a.log.Error("create user error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusInternalServerError, Body: "Internal server error"}, a.log)
		return
	}

	httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusOK, Body: nil}, a.log)
}

func (a *Api) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		a.log.Error("Logout error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusBadRequest, Body: models.ErrorResponse{Error: "Bad request"}}, a.log)
		return
	}

	err = a.session.KillSession(r.Context(), cookie.Value)
	if err != nil {
		a.log.Error("Logout error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusInternalServerError, Body: models.ErrorResponse{Error: "Internal server error"}}, a.log)
		return
	}

	cookie.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, cookie)

	httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusOK, Body: nil}, a.log)
}

func (a *Api) AuthAccept(w http.ResponseWriter, r *http.Request) {
	var authorized bool

	session, err := r.Cookie("session_id")
	if err == nil && session != nil {
		authorized, err = a.session.FindActiveSession(r.Context(), session.Value)
		if err != nil {
			a.log.Error("AuthAccept error: ", err.Error())
			httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusUnauthorized, Body: models.ErrorResponse{Error: "Not authorized"}}, a.log)
			return
		}
	}

	if !authorized {
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusUnauthorized, Body: models.ErrorResponse{Error: "Not authorized"}}, a.log)
		return
	}

	login, err := a.session.GetUserName(r.Context(), session.Value)
	if err != nil {
		a.log.Error("auth accept error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusInternalServerError, Body: models.ErrorResponse{Error: "Internal server error"}}, a.log)
		return
	}

	httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusOK, Body: models.AuthCheckResponse{Login: login}}, a.log)
}

func (a *Api) GetEmployees(w http.ResponseWriter, r *http.Request) {
	offset, err := strconv.ParseUint(r.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		offset = 0
	}

	limit, err := strconv.ParseUint(r.URL.Query().Get("limit"), 10, 64)
	if err != nil {
		limit = 8
	}

	emps, err := a.profile.GetEmployees(r.Context(), offset, limit)
	if err != nil {
		a.log.Error("Get employees error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusInternalServerError, Body: models.ErrorResponse{Error: "Internal server error"}}, a.log)
		return
	}

	httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusOK, Body: emps}, a.log)
}

func (a *Api) BirthdaySub(w http.ResponseWriter, r *http.Request) {
	var request models.SubRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.log.Error("BirthdaySub error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusBadRequest, Body: models.ErrorResponse{Error: "Bab request"}}, a.log)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		a.log.Error("Error unmarshalling request: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusBadRequest, Body: models.ErrorResponse{Error: "Bab request"}}, a.log)
		return
	}

	request.UserFromId = r.Context().Value(middleware.UserIDKey).(uint64)
	res, err := a.sub.BirthdaySub(r.Context(), request.UserFromId, request.UserToId)
	if err != nil {
		if err.Error() == errs.ErrDuplicateSub.Error() {
			httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusConflict, Body: models.ErrorResponse{Error: "Already exist"}}, a.log)
			return
		}

		a.log.Error("Birthday sub error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusInternalServerError, Body: models.ErrorResponse{Error: "Internal server error"}}, a.log)
		return
	}

	if !res {
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusNotFound, Body: models.ErrorResponse{Error: "Not found"}}, a.log)
		return
	}

	httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusOK, Body: nil}, a.log)
}

func (a *Api) BirthdayUnSub(w http.ResponseWriter, r *http.Request) {
	var request models.UnSubRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.log.Error("BirthdayUnSub error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusBadRequest, Body: models.ErrorResponse{Error: "Bad request"}}, a.log)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		a.log.Error("Error unmarshalling request: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusBadRequest, Body: models.ErrorResponse{Error: "Bad request"}}, a.log)
		return
	}

	request.UserFromId = r.Context().Value(middleware.UserIDKey).(uint64)
	_, err = a.sub.BirthdayUnSub(r.Context(), request.UserFromId, request.UserToId)
	if err != nil {
		if err.Error() == errs.ErrNotFound.Error() {
			httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusNotFound, Body: models.ErrorResponse{Error: "Not found"}}, a.log)
			return
		}

		a.log.Error("Birthday sub error: ", err.Error())
		httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusInternalServerError, Body: models.ErrorResponse{Error: "Internal server error"}}, a.log)
		return
	}

	httpResponse.SendResponse(w, r, &models.Response{Status: http.StatusOK, Body: nil}, a.log)
}
