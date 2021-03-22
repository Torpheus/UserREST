package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type user struct {
	Name     string
	Email    string
	Password []byte
}

var dbLock sync.RWMutex
var userDb []user

func main() {
	userDb = append(userDb, user{"test", "user", nil})
	http.HandleFunc("/users/", userAPI)
	err := http.ListenAndServe("0.0.0.0:8477", nil)
	if err != nil {
		log.Fatalln("Failed to start server")
	}
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
	decoder := json.NewDecoder(req.Body)
	var newUser user
	dbLock.Lock()
	err := decoder.Decode(&newUser)
	dbLock.Unlock()
	if err != nil {
		log.Printf("Server \"PUT /users/\", failed to decode json\n")
		return
	}
	userDb = append(userDb, newUser)
	return
}

func getUser(path string, w http.ResponseWriter) {
	id, err := strconv.Atoi(path)
	if err != nil || len(userDb) < id {
		log.Printf("Serving \"GET /users/%s\", failed to convert path to id\n", path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encode := json.NewEncoder(w)
	dbLock.RLock()
	err = encode.Encode(userDb[id])
	dbLock.RUnlock()
	if err != nil {
		log.Printf("Serving \"GET /users/%s\", failed to json encode user\n", path)
	}
	return
}

func getAll(w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	dbLock.RLock()
	err := encoder.Encode(userDb)
	dbLock.RUnlock()
	if err != nil {
		log.Println("Serving \"GET /users/\", json encode failed")
	}
	return
}
