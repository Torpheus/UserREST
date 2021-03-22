package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type user struct {
	Id       int
	Name     string
	Email    string
	Password string
}

var database *sql.DB
var queryUsers *sql.Stmt
var queryUser *sql.Stmt
var insertUser *sql.Stmt

func main() {

	database = initDatabase()

	http.HandleFunc("/users/", userAPI)
	err := http.ListenAndServe("0.0.0.0:8477", nil)
	if err != nil {
		log.Fatalln("Failed to start server")
	}

	_ = queryUsers.Close()
	_ = queryUser.Close()
	_ = insertUser.Close()

	_ = database.Close()
}

func initDatabase() *sql.DB {
	db, err := sql.Open("mysql", "root:password@tcp(172.17.0.2)/users")
	if err != nil || db.Ping() != nil {
		log.Fatalln("Failed to open database")
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	queryUsers, err = db.Prepare("SELECT * FROM users")
	if err != nil {
		log.Fatalln("Failed to prepare statement \"SELECT * FROM users\"")
	}
	queryUser, err = db.Prepare("SELECT * FROM users WHERE Id = ?")
	if err != nil {
		log.Fatalln("Failed to prepare statement \"SELECT * FROM users WHERE Id = ?\"")
	}
	insertUser, err = db.Prepare("INSERT INTO users(Name, Email, Password) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatalln("Failed to prepare statement \"INSERT INTO users(Name, Email, Password) VALUES (?, ?, ?)\"")
	}
	return db
}

func userAPI(w http.ResponseWriter, req *http.Request) {
	path := strings.Replace(req.URL.Path, "/users/", "", 1)

	switch req.Method {
	case http.MethodGet:
		if len(path) == 0 {
			getAll(w)
		} else {
			getUser(path, w)
		}
	case http.MethodPut:
		addUser(req)
	}
}

func addUser(req *http.Request) {
	if req.Header.Get("Content-Type") != "application/json" {
		log.Printf("Server \"PUT /users/\", tried to put content type %q\n", req.Header.Get("Content-Type"))
		return
	}
	var newUser user

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newUser)

	if err != nil {
		log.Printf("Server \"PUT /users/\", failed to decode json\n")
		return
	}

	_, err = insertUser.Exec(newUser.Name, newUser.Email, newUser.Password)
	if err != nil {
		log.Printf("Inserting of user %v into database failed with %v\n", newUser, err)
	}
}

func getUser(path string, w http.ResponseWriter) {
	id, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("Serving \"GET /users/%s\", failed to convert path to id\n", path)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var selectedUser user
	row := queryUser.QueryRow(id)
	err = row.Scan(&selectedUser.Id, &selectedUser.Name, &selectedUser.Email, &selectedUser.Password)
	if err != nil {
		log.Printf("Fetching user id %d from database failed", id)
		return
	}
	encode := json.NewEncoder(w)
	err = encode.Encode(selectedUser)

	if err != nil {
		log.Printf("Serving \"GET /users/%s\", failed to json encode user %d\n", path, id)
	}
}

func getAll(w http.ResponseWriter) {
	rows, err := queryUsers.Query()
	if err != nil {
		log.Printf("Querying of all users failed with %v\n", err)
		return
	}

	var users []user
	for rows.Next() {
		var user user
		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password)
		if err != nil {
			log.Printf("%v\n", err)
			return
		}
		users = append(users, user)
	}
	if rows.Err() != nil {
		log.Printf("%v\n", rows.Err())
		return
	}

	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(users)
	if err != nil {
		log.Println("Serving \"GET /users/\", json encode failed")
	}
}
