package main

import (
	"encoding/json"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type user struct {
	Id       int
	Name     string
	Email    string
	Password string
}

var listenAddress = flag.String("addr",
	"127.0.0.1",
	"Address on which the server listens for requests")

var listenPort = flag.Int("port",
	8444,
	"Port on which the server listens for requests")

var databasePort = flag.Int("dbport",
	3306,
	"Port for reaching the database")

var databaseAddress = flag.String("dbaddr",
	"127.0.0.1",
	"Address of the database server")

var databaseUser = flag.String("dbuser",
	"root",
	"User for database login")

var databasePassword = flag.String("dbpass",
	"",
	"Password for database login")
var databaseName = flag.String("dbname",
	"users",
	"Name of the user database")

func main() {
	flag.Parse()

	database = initDatabase()

	http.HandleFunc("/users/", userAPI)
	err := http.ListenAndServe(*listenAddress+":"+strconv.Itoa(*listenPort), nil)
	if err != nil {
		log.Fatalln("Failed to start server")
	}

	_ = queryUsers.Close()
	_ = queryUser.Close()
	_ = insertUser.Close()

	_ = database.Close()
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

	newUser.insert()
}

func getUser(path string, w http.ResponseWriter) {
	id, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("Serving \"GET /users/%s\", failed to convert path to id\n", path)
		return
	}

	var selectedUser user
	selectedUser.Id = id
	selectedUser.query()

	w.Header().Set("Content-Type", "application/json")

	encode := json.NewEncoder(w)
	err = encode.Encode(selectedUser)

	if err != nil {
		log.Printf("Serving \"GET /users/%s\", failed to json encode user %d\n", path, id)
	}
}

func getAll(w http.ResponseWriter) {
	users := queryAll()

	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	err := encoder.Encode(users)
	if err != nil {
		log.Println("Serving \"GET /users/\", json encode failed")
	}
}
