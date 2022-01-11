package main

import (
	"fmt"
	"bytes"
	"time"
	"io/ioutil"
	"net/smtp"
	"net/http"
	"encoding/json"
	"database/sql"
	"text/template"
	"github.com/gorilla/mux"
	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
)

type SignUPBody struct {
	Email, Password string
}

type User struct {
	Id int
	Email string
}

type PostBody struct {
	Email, Password string
}

type Payload struct {
	Email, Password string
	jwt.StandardClaims
}

type Course struct {
	Id int 
	Name string
	Price float64
}

var SQL_INSERT_USER = `
	insert into users (
		email, password
	) values ($1, $2)
	returning
		user_id,
		email
`

var jwtKey = []byte("gophers")

func PostSignUpCtrl(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	encoder := json.NewEncoder(w)

	signUp := SignUPBody {}

	json.Unmarshal(body, &signUp)

	db, err := sql.Open("postgres", DB_CONFIG)

	defer db.Close()

	if err != nil {
		panic(err)
	}

	user := User {}

	err = db.QueryRow(
		SQL_INSERT_USER,
		signUp.Email,
		signUp.Password,
	).Scan(
		&user.Id,
		&user.Email,
	)

	var uuid string

	err = db.QueryRow(
		`select id from activation where user_id = $1`,
		user.Id,
	).Scan(&uuid)

	auth := smtp.PlainAuth(
		"",
		"goguruh01@gmail.com",
		"Qwertyu!op",
		"smtp.gmail.com",
	)

	var buffer bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	buffer.Write([]byte(fmt.Sprintf("Subject: Welcome to our site! \n%s\n\n", mimeHeaders)))

	t, err := template.ParseFiles("mail-template.html")

	t.Execute(&buffer, struct { Email, UUID string}{
		Email: user.Email,
		UUID: uuid,
	})

	err = smtp.SendMail("smtp.gmail.com:587", auth, "goguruh01@gmail.com", []string { user.Email }, buffer.Bytes())

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")

	encoder.Encode(user)
}


func VerifyCtrl(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	db, err := sql.Open("postgres", DB_CONFIG)

	defer db.Close()

	row, err := db.Exec(
		`update users set activated_at = current_timestamp
		where user_id = (select user_id from activation where id = $1)`,
		vars["uuid"],
	)

	if err != nil {
		panic(err)
	}

	affected, _ := row.RowsAffected()

	if affected > 0 {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}


func LoginCtrl(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)

	login := PostBody{}

	json.Unmarshal(body, &login)

	db, err := sql.Open("postgres", DB_CONFIG)

	defer db.Close()

	if err != nil {
		panic(err)
	}

	var id int 
	// var activated string

	err = db.QueryRow(
		"select user_id from users where email = $1 and password = $2",
		login.Email,
		login.Password,
	).Scan(&id)

	if err != nil {
		panic(err)
	}

	if id != 0 {

		expirationTime := time.Now().Add(500 * time.Second)

		payload := Payload {
			Email: login.Email,
			Password: login.Password,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

		tokenString, err := token.SignedString(jwtKey)

		if err != nil {
			panic(err)
		}

		w.Write([]byte(tokenString))

	} else {
		w.Write([]byte("Wrong username or password"))
	}

}


func GetCoursesCtrl(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("token")

	encoder := json.NewEncoder(w)

	payload := &Payload{}

	tkn, err := jwt.ParseWithClaims(token, payload, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	db, err := sql.Open("postgres", DB_CONFIG)

	defer db.Close()

	if err != nil {
		panic(err)
	}

	var courses []Course

	rows, err := db.Query(
		`select 
			course_id,
			name,
			price
		from 
			courses
			`,
	)

	defer db.Close()

	for rows.Next() {

		var course Course

		err = rows.Scan(&course.Id, &course.Name, &course.Price); 

		if err != nil {
            panic(err)
        }
        	
        courses = append(courses, course)
	}	

	encoder.Encode(courses)
}

