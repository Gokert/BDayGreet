package delivery

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"time"
	"vk-rest/pkg/middleware"
	"vk-rest/pkg/models"
	httpResponse "vk-rest/pkg/response"
	"vk-rest/usecase"
)

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
	api.mx.Handle("/authcheck", md.MethodCheck(http.HandlerFunc(api.AuthAccept), http.MethodDelete))
	api.mx.Handle("/api/v1/employees/list", md.AuthCheck(md.MethodCheck(http.HandlerFunc(api.GetEmployees), http.MethodGet)))
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
	response := models.Response{Status: http.StatusOK, Body: nil}
	var request models.SigninRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.log.Error("Signin error: ", err.Error())
		response.Status = http.StatusBadRequest
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		a.log.Error("Signin error: ", err.Error())
		response.Status = http.StatusBadRequest
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	_, found, err := a.profile.FindUserAccount(r.Context(), request.Login, request.Password)
	if err != nil {
		a.log.Error("Signin error: ", err.Error())
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	if !found {
		response.Status = http.StatusUnauthorized
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

	httpResponse.SendResponse(w, r, &response, a.log)
}

func (a *Api) Signup(w http.ResponseWriter, r *http.Request) {
	response := models.Response{Status: http.StatusOK, Body: nil}
	var request models.SignupRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Status = http.StatusBadRequest
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	found, err := a.profile.FindUserByLogin(r.Context(), request.Login)
	if err != nil {
		a.log.Error("Signup error: ", err.Error())
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	if found {
		response.Status = http.StatusConflict
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	err = a.profile.CreateUserAccount(r.Context(), request.Login, request.Password)
	if err != nil {
		a.log.Error("create user error: ", err.Error())
		response.Status = http.StatusBadRequest
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	httpResponse.SendResponse(w, r, &response, a.log)
}

func (a *Api) Logout(w http.ResponseWriter, r *http.Request) {
	response := models.Response{Status: http.StatusOK, Body: nil}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		response.Status = http.StatusBadRequest
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	err = a.session.KillSession(r.Context(), cookie.Value)
	if err != nil {
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	cookie.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, cookie)

	httpResponse.SendResponse(w, r, &response, a.log)
}

func (a *Api) AuthAccept(w http.ResponseWriter, r *http.Request) {
	response := models.Response{Status: http.StatusOK, Body: nil}
	var authorized bool

	session, err := r.Cookie("session_id")
	if err == nil && session != nil {
		authorized, _ = a.session.FindActiveSession(r.Context(), session.Value)
	}
	a.log.Warning("API", authorized)
	if !authorized {
		response.Status = http.StatusUnauthorized
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	login, err := a.session.GetUserName(r.Context(), session.Value)
	if err != nil {
		a.log.Error("auth accept error: ", err.Error())
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	response.Body = models.AuthCheckResponse{
		Login: login,
	}

	httpResponse.SendResponse(w, r, &response, a.log)
}

func (a *Api) GetEmployees(w http.ResponseWriter, r *http.Request) {
	response := models.Response{Status: http.StatusOK, Body: nil}

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
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	response.Body = emps
	httpResponse.SendResponse(w, r, &response, a.log)

}

func (a *Api) BirthdaySub(w http.ResponseWriter, r *http.Request) {
	response := models.Response{Status: http.StatusOK, Body: nil}
	var request models.SubRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Status = http.StatusBadRequest
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	res, err := a.sub.BirthdaySub(r.Context(), request.UserFromId, request.UserToId)
	if err != nil {
		a.log.Error("Birthday sub error: ", err.Error())
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	if !res {
		response.Status = http.StatusNotFound
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	httpResponse.SendResponse(w, r, &response, a.log)
}

func (a *Api) BirthdayUnSub(w http.ResponseWriter, r *http.Request) {
	response := models.Response{Status: http.StatusOK, Body: nil}
	var request models.UnSubRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Status = http.StatusBadRequest
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	res, err := a.sub.BirthdayUnSub(r.Context(), request.UserFromId, request.UserToId)
	if err != nil {
		a.log.Error("Birthday sub error: ", err.Error())
		response.Status = http.StatusInternalServerError
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	if !res {
		response.Status = http.StatusNotFound
		httpResponse.SendResponse(w, r, &response, a.log)
		return
	}

	httpResponse.SendResponse(w, r, &response, a.log)
}
