package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"runtime"

	"github.com/julienschmidt/httprouter"
)

func init() {

	router := httprouter.New()

	//router.GET("/user/:userId", ErrorHandler(LogHandler(getUserDetails)))

	router.GET("/user", ErrorHandler(LogHandler(HomeHandler)))

	router.POST("/login", ErrorHandler(LogHandler(LoginHandler)))

	router.GET("/admin", ErrorHandler(LogHandler(AdminHandler)))
	router.POST("/createuser", ErrorHandler(LogHandler(createUserHandler)))

	router.POST("/transfermoney", ErrorHandler(LogHandler(TransfermoneyHandler)))

	//to add devices through rest api

	router.GET("/_ah/start", StartHandler)
	router.GET("/_ah/stop", StopHandler)
	http.Handle("/", router)

	loadTemplates()

}

type Auth struct {
	Username string
}

var templates map[string]*template.Template

func loadTemplates() {
	templates = make(map[string]*template.Template, 5)
	files := []string{"login.html", "user.html", "admin.html"}

	for _, t := range files {
		templates[t] = template.Must(template.ParseFiles("base.html", t))
	}
}

func PanicHandler(funcName string) {
	if r := recover(); r != nil {
		//TODO:generate a issue number and log the details here. maybe take w, r from handler and log params from there also for debugging later.
		fmt.Println("Recovered from panic in "+funcName+"\nPanic: ", r)
	}
}

type AppHandler func(http.ResponseWriter, *http.Request, httprouter.Params) *HttpError
type AppAuthHandler func(http.ResponseWriter, *http.Request, httprouter.Params, *Auth) *HttpError

func ErrorHandler(a AppHandler) httprouter.Handle { //respond to user with error in requested encoding
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer PanicHandler(runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name())
		if e := a(w, r, p); e != nil {

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(e.Code)

			json.NewEncoder(w).Encode(struct {
				*HttpError
				IsError bool `json:"error"`
			}{
				HttpError: e,
				IsError:   true,
			})
		}
	}
}

func PrintSuccessJson(w http.ResponseWriter, j []byte) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization,Token")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if bytes.Equal(j, []byte("null")) {
		fmt.Fprintf(w, "%s", "{}")
	} else {
		fmt.Fprintf(w, "%s", j)
	}
}

func VersionHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HttpError {
	fmt.Fprintf(w, "Version:%s", os.Getenv("CURRENT_VERSION_ID"))
	return nil
}

func StartHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "Success")
}
func StopHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "Success")
}
