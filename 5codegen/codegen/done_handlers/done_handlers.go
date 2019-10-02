package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"strconv"
)

type myResponse struct {
	Error    string      `json:"error"`
	Response interface{} `json:"response,omitempty"`
}

func (h *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//ответ
	MyResp := new(myResponse)

	//получаем логин
	var Params ProfileParams

	switch r.Method {
	case http.MethodGet:
		Params.Login = r.URL.Query().Get("login")
	case http.MethodPost:
		r.ParseForm()
		Params.Login = r.Form.Get("login")
	}

	if Params.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "login must me not empty"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}

	user, err := h.Profile(ctx, Params)

	if user != nil {
		MyResp.Response = user
	}

	if err != nil {
		MyResp.Error = err.Error()
		if _, ok := err.(ApiError); ok {
			w.WriteHeader(http.StatusNotFound)
			mr, _ := json.Marshal(MyResp)
			w.Write(mr)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}

	mr, err := json.Marshal(MyResp)

	if err != nil {
		fmt.Println("Cant pack json: ", err)
	}
	w.Write(mr)
}

func (h *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var Params CreateParams

	MyResp := new(myResponse)



	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusNotAcceptable)
		MyResp.Error = "bad method"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	case http.MethodPost:
		r.ParseForm()

		var err error

		Params.Login = r.Form.Get("login")
		Params.Age, err = strconv.Atoi(r.Form.Get("age"))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			MyResp.Error = "age must be int"
			mr, _ := json.Marshal(MyResp)
			w.Write(mr)
			return

		}

		Params.Name = r.Form.Get("full_name")
		Params.Status = r.Form.Get("status")
	}

	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		MyResp.Error = "unauthorized"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}

	if Params.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "login must me not empty"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}

	if len(Params.Login) <= 10 {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "login len must be >= 10"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}

	if Params.Age < 0 {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "age must be >= 0"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}

	if Params.Age > 128 {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "age must be <= 128"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}
	if Params.Status == "" {
		Params.Status = "user"
	}

	if (Params.Status != "user") && (Params.Status != "moderator") && (Params.Status != "admin") {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "status must be one of [user, moderator, admin]"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}

	NewUser, err := h.Create(ctx, Params)

	if err, ok := err.(ApiError); ok {
		w.WriteHeader(err.HTTPStatus)
	}

	if NewUser != nil {
		MyResp.Response = NewUser
	}
	if err != nil {
		MyResp.Error = err.Error()
		if _, ok := err.(ApiError); ok {
			w.WriteHeader(http.StatusNotFound)
			mr, _ := json.Marshal(MyResp)
			w.Write(mr)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
	mr, err := json.Marshal(MyResp)
	if err != nil {
		fmt.Println("Cant pack json:", err)
	}
	w.Write(mr)
}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {
	case "/user/profile":
		h.handlerProfile(w, r)
	case "/user/create":
		h.handlerCreate(w, r)

	default:
		MyResp := new(myResponse)
		MyResp.Error = "unknown method"
		mr, err := json.Marshal(MyResp)
		if err != nil {
			fmt.Println("Cant pack json:", err)
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write(mr)
	}
}

func (h *OtherApi) handlerProfile(w http.ResponseWriter, r *http.Request) {}
func (h *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request)  {}
func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {

	case "/user/create":
		h.handlerCreate(w, r)

	default:
		MyResp := new(myResponse)
		MyResp.Error = "unknown method"
		mr, err := json.Marshal(MyResp)
		if err != nil {
			fmt.Println("Cant pack json:", err)
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write(mr)
	}
}
