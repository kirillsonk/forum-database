package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	initDatabase()
	router()
}

func initDatabase() {
	const (
		host     = "localhost"
		port     = 5432
		user     = "ksonk"
		password = "k123"
		dbname   = "forumdb"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	initDB, err := ioutil.ReadFile("./sql/DBtables.sql")
	_, err = db.Exec(string(initDB))

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}

func router() {
	router := mux.NewRouter()

	router.HandleFunc("/api/forum/create", createForum)
	router.HandleFunc("/api/forum/{slug}/create", createThread)
	router.HandleFunc("/api/forum/{slug}/details", forumDetails)
	router.HandleFunc("/api/forum/{slug}/threads", forumThreads)
	router.HandleFunc("/api/forum/{slug}/users", forumUsers)
	router.HandleFunc("/api/post/{id}/details", postDetails)
	router.HandleFunc("/api/service/clear", serviceClear)
	router.HandleFunc("/api/service/status", serviceStatus)
	router.HandleFunc("/api/thread/{slug_or_id}/create", postsCreate)
	router.HandleFunc("/api/thread/{slug_or_id}/details", threadDetails)
	router.HandleFunc("/api/thread/{slug_or_id}/posts", threadPosts)
	router.HandleFunc("/api/thread/{slug_or_id}/vote", threadVote)
	router.HandleFunc("/api/user/{nickname}/create", userCreate)
	router.HandleFunc("/api/user/{nickname}/profile", userProfile)

	http.Handle("/", router)
	fmt.Println("Listen server on port: 5000")
	http.ListenAndServe(":5000", nil)
}
