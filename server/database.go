package main

import (
	"database/sql"
	"log"
	"strconv"
	"time"
)

var database *sql.DB
var queryUsers *sql.Stmt
var queryUser *sql.Stmt
var insertUser *sql.Stmt

func initDatabase() *sql.DB {
	db, err := sql.Open("mysql",
		*databaseUser+":"+*databasePassword+"@tcp("+*databaseAddress+":"+
			strconv.Itoa(*databasePort)+")/"+*databaseName)
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

func (u user) insert() {
	_, err := insertUser.Exec(u.Name, u.Email, u.Password)
	if err != nil {
		log.Printf("Inserting of user %v into database failed with %v\n", u, err)
	}
}

func (u *user) query() {
	row := queryUser.QueryRow(u.Id)
	err := row.Scan(&u.Id, &u.Name, &u.Email, &u.Password)
	if err != nil {
		log.Printf("Fetching user id %d from database failed", u.Id)
		return
	}
}

func queryAll() (users []user) {
	rows, err := queryUsers.Query()
	if err != nil {
		log.Printf("Querying of all users failed with %v\n", err)
		return nil
	}

	for rows.Next() {
		var user user
		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password)
		if err != nil {
			log.Printf("%v\n", err)
			return nil
		}
		users = append(users, user)
	}
	if rows.Err() != nil {
		log.Printf("%v\n", rows.Err())
		return nil
	}
	return
}
