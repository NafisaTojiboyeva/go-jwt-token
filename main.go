package main

import (
	"net/http"
	"github.com/gorilla/mux"
)


func main() {

	r := mux.NewRouter()

	r.HandleFunc("/signup", PostSignUpCtrl).Methods("POST")
	r.HandleFunc("/verify/{uuid}", VerifyCtrl)

	r.HandleFunc("/login", LoginCtrl).Methods("POST")
	r.HandleFunc("/courses", GetCoursesCtrl).Methods("GET")


	http.ListenAndServe(":8080", r)
}
