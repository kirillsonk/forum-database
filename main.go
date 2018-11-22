package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	// "github.com/bozaro/tech-db-forum/generated/models"

	models "./models"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	const (
		host     = "localhost"
		user     = "ksonk"
		password = "k123"
		dbname   = "forumdb"
	)
	var err error
	psqlInfo := fmt.Sprintf("host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, user, password, dbname)

	db, err = sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}

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

func main() {
	router()
	defer db.Close()
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
	// router.HandleFunc("/api/thread/{slug_or_id}/create", postsCreate)
	router.HandleFunc("/api/thread/{slug_or_id}/details", threadDetails)
	router.HandleFunc("/api/thread/{slug_or_id}/posts", threadPosts)
	router.HandleFunc("/api/thread/{slug_or_id}/vote", threadVote)
	router.HandleFunc("/api/user/{nickname}/create", userCreate)
	router.HandleFunc("/api/user/{nickname}/profile", userProfile)

	http.Handle("/", router)
	fmt.Println("Listen server on port: 5000")
	http.ListenAndServe(":5000", nil)
}

func createForum(w http.ResponseWriter, r *http.Request) { //POST +
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		reqBody, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var forum models.Forum
		var newForum models.Forum
		var oldForum models.Forum

		err = json.Unmarshal(reqBody, &forum)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = db.QueryRow("SELECT nickname FROM Users WHERE nickname=$1", forum.User).Scan(&forum.User)
		if err != nil {
			if err == sql.ErrNoRows {
				var e models.Error
				e.Message = "Can't find user with nickname: " + forum.User + "\n"
				resData, _ := json.Marshal(e)
				w.WriteHeader(http.StatusNotFound)
				w.Write(resData)
				return
			}
		}

		err = db.QueryRow("INSERT INTO Forums (slug, title, author) VALUES ($1, $2, $3) RETURNING slug, title, author;",
			forum.Slug,
			forum.Title,
			forum.User).Scan(
			// &newForum.Posts,
			&newForum.Slug,
			// &newForum.Threads,
			&newForum.Title,
			&newForum.User)

		// fmt.Println(forum)
		// fmt.Println("----------------------")
		// fmt.Println(newForum)
		// fmt.Println("----------------------")

		if err != nil {
			err = db.QueryRow("SELECT slug, title, author FROM Forums WHERE author=$1", forum.User).
				Scan(&oldForum.Slug,
					&oldForum.Title,
					&oldForum.User)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resData, _ := json.Marshal(oldForum)
			w.WriteHeader(http.StatusConflict)
			w.Write(resData)
			return
		}
		resData, _ := json.Marshal(newForum)
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
		return
	}
	return
}

func createThread(w http.ResponseWriter, r *http.Request) { //POST + (исправлять)
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer r.Body.Close()

		var thread models.Thread
		var newThread models.Thread

		err = json.Unmarshal(reqBody, &thread)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		args := mux.Vars(r)
		Slug := args["slug"]
		thread.Slug = Slug

		err = db.QueryRow("INSERT INTO Threads (author, created, message, title, slug) VALUES ($1, $2, $3, $4, $5) RETURNING *",
			thread.Author,
			thread.Created,
			thread.Message,
			thread.Title,
			thread.Slug).
			Scan(&newThread.Author,
				&newThread.Created,
				&newThread.Forum,
				&newThread.Id,
				&newThread.Message,
				&newThread.Slug,
				&newThread.Title,
				&newThread.Votes)

		// if err != nil { //404
		// 	fmt.Println(err.Error())
		// 	if err == sql.ErrNoRows {
		// 		var e models.Error
		// 		e.Message = "Can't find user with id " + thread.Author
		// 		resData, _ := json.Marshal(e)
		// 		w.WriteHeader(http.StatusNotFound)
		// 		w.Write(resData)
		// 		return
		// 	}
		// 	fmt.Println(" ----------------------------------------")

		// 	//409
		// 	err = db.QueryRow("SELECT * FROM Threads WHERE title=$1 OR slug=$2", thread.Title, thread.Forum).
		// 		Scan(&thread.Author,
		// 			&thread.Created,
		// 			&thread.Forum,
		// 			&thread.Id,
		// 			&thread.Message,
		// 			&thread.Slug,
		// 			&thread.Title,
		// 			&thread.Votes,
		// 		)
		// 	if err != nil {
		// 		w.WriteHeader(http.StatusInternalServerError)
		// 		return

		// 	}
		// 	resData, _ := json.Marshal(thread)
		// 	w.WriteHeader(http.StatusConflict)
		// 	w.Write(resData)
		// 	return
		// }

		resData, _ := json.Marshal(newThread)
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
		return
	}
	return
}

func forumDetails(w http.ResponseWriter, r *http.Request) { //GET +  (вероятно, неправильно написан)
	if r.Method == http.MethodGet {
		w.Header().Set("content-type", "application/json")

		args := mux.Vars(r)
		Slug := args["slug"]

		var forum models.Forum

		// fmt.Println(Slug)
		// fmt.Println("------------------")

		err := db.QueryRow("SELECT slug, title, author FROM Forums WHERE slug = $1;", Slug).
			// err := db.QueryRow("SELECT * FROM Forums WHERE slug = $1;", Slug).
			Scan(
				// &forum.Posts,
				&forum.Slug,
				// &forum.Threads,
				&forum.Title,
				&forum.User)

		// fmt.Println(forum)
		// fmt.Println("------------------")
		if err != nil {
			fmt.Println(err.Error())
			var e models.Error
			e.Message = "Can't find user with slug " + Slug
			resData, _ := json.Marshal(e)

			w.WriteHeader(http.StatusNotFound)
			w.Write(resData)
			return
		}

		resData, _ := json.Marshal(forum)
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
		return
	}
	return
}

func forumThreads(w http.ResponseWriter, r *http.Request) { //GET - (Sort) сложна оптимизировать
	if r.Method == http.MethodGet {

		limitVal := r.URL.Query().Get("limit")
		sinceVal := r.URL.Query().Get("since")
		descVal := r.URL.Query().Get("desc")

		var limit = false
		var since = false
		var desc = false

		if limitVal != "" {
			limit = true
		}
		if sinceVal != "" {
			since = true
		}
		if descVal == "true" {
			desc = true
		}

		vars := mux.Vars(r)
		slug := vars["slug"]

		var rows *sql.Rows
		var err error

		if limit && !since && !desc {
			rows, err = db.Query("SELECT * FROM Threads WHERE Forum = $1 ORDER BY created LIMIT $2;", slug, limitVal)
		} else if since && !limit && !desc {
			rows, err = db.Query("SELECT * FROM Threads WHERE Forum = $1 AND created <= $2 ORDER BY created;", slug, sinceVal)
		} else if limit && since && !desc {
			rows, err = db.Query("SELECT * FROM Threads WHERE Forum = $1 AND created >= $2 ORDER BY created LIMIT $3;", slug, sinceVal, limitVal)
		} else if limit && !since && desc {
			rows, err = db.Query("SELECT * FROM Threads WHERE Forum = $1 ORDER BY created DESC LIMIT $2;", slug, limitVal)
		} else if since && !limit && desc {
			rows, err = db.Query("SELECT * FROM Threads WHERE Forum = $1 AND created <= $2 ORDER BY created DESC;", slug, sinceVal)
		} else if limit && since && desc {
			rows, err = db.Query("SELECT * FROM Threads WHERE Forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3;", slug, sinceVal, limitVal)
		} else if limit && since && !desc {
			rows, err = db.Query("SELECT * FROM Threads WHERE Forum = $1 AND created >= $2 ORDER BY created LIMIT $3;", slug, sinceVal, limitVal)
		} else if !limit && !since && !desc {
			rows, err = db.Query("SELECT * FROM Threads WHERE Forum = $1 ORDER BY created;", slug)
		} else {
			rows, err = db.Query("SELECT * FROM Threads WHERE Forum = $1 ORDER BY created;", slug)
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		thrs := make([]models.Thread, 0)

		for rows.Next() {
			// thr := Thread{}
			var thr models.Thread
			err := rows.Scan(&thr.Id, &thr.Author, &thr.Created, &thr.Forum, &thr.Message, &thr.Slug, &thr.Title, &thr.Votes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			thrs = append(thrs, thr)
		}
		if getForum(slug) == nil {
			// e := new(Error)
			var e models.Error
			e.Message = "Can't find forum with slug " + slug + "\n"
			resp, _ := json.Marshal(e)
			w.Header().Set("content-type", "application/json")

			w.WriteHeader(http.StatusNotFound)
			w.Write(resp)
			return
		}

		resp, _ := json.Marshal(thrs)
		w.Header().Set("content-type", "application/json")

		w.Write(resp)

		return
	}
	return
}

func forumUsers(w http.ResponseWriter, r *http.Request) { //GET - (Sort) Сделать (без рекурсии)
	if r.Method == http.MethodGet {
		limitVal := r.URL.Query().Get("limit")
		sinceVal := r.URL.Query().Get("since")
		descVal := r.URL.Query().Get("desc")

		var limit = false
		var since = false
		var desc = false

		if limitVal != "" {
			limit = true
		}
		if sinceVal != "" {
			since = true
		}
		if descVal == "true" {
			desc = true
		}

		var rows *sql.Rows
		var err error

		vars := mux.Vars(r)
		slug := vars["slug"]

		frm := getForum(slug)

		if frm == nil {
			// e := new(Error)
			var e models.Error
			e.Message = "Can't find forum with slug " + slug + "\n"
			resp, _ := json.Marshal(e)
			w.Header().Set("content-type", "application/json")

			w.WriteHeader(http.StatusNotFound)
			w.Write(resp)
			return
		}

		if !limit && !since && !desc {
			rows, err = db.Query("SELECT * FROM Users WHERE nickname IN (SELECT author FROM Threads WHERE forum=$1 GROUP BY author) OR nickname IN (SELECT author FROM Posts WHERE forum=$1 GROUP BY author) ORDER BY nickname ASC;", slug)
		} else if !limit && !since && desc {
			rows, err = db.Query("SELECT * FROM Users WHERE nickname IN (SELECT author FROM Threads WHERE forum=$1 GROUP BY author) OR nickname IN (SELECT author FROM Posts WHERE forum=$1 GROUP BY author) ORDER BY nickname DESC;", slug)
		} else if !limit && since && !desc {
			rows, err = db.Query("SELECT * FROM Users WHERE nickname IN (SELECT author FROM Threads WHERE forum=$1 AND author>$2 GROUP BY author) OR nickname IN (SELECT author FROM Posts WHERE forum=$1 AND author>$2 GROUP BY author) AND nickname>$2 ORDER BY nickname ASC;", slug, sinceVal)
		} else if !limit && since && desc {
			rows, err = db.Query("SELECT * FROM Users WHERE nickname IN (SELECT author FROM Threads WHERE forum=$1 AND author<$2 GROUP BY author) OR nickname IN (SELECT author FROM Posts WHERE forum=$1 AND author<$2 GROUP BY author) AND nickname<$2 ORDER BY nickname DESC;", slug, sinceVal)
		} else if limit && !since && !desc {
			rows, err = db.Query("SELECT * FROM Users WHERE nickname IN (SELECT author FROM Threads WHERE forum=$1 GROUP BY author) OR nickname IN (SELECT author FROM Posts WHERE forum=$1 GROUP BY author) ORDER BY nickname ASC LIMIT $2;", slug, limitVal)
		} else if limit && !since && desc {
			rows, err = db.Query("SELECT * FROM Users WHERE nickname IN (SELECT author FROM Threads WHERE forum=$1 GROUP BY author) OR nickname IN (SELECT author FROM Posts WHERE forum=$1 GROUP BY author) ORDER BY nickname DESC LIMIT $2;", slug, limitVal)
		} else if limit && since && !desc {
			rows, err = db.Query("SELECT * FROM Users WHERE nickname IN (SELECT author FROM Threads WHERE forum=$1 AND author>$2 GROUP BY author) OR nickname IN (SELECT author FROM Posts WHERE forum=$1 AND author>$2 GROUP BY author) AND nickname>$2 ORDER BY nickname ASC LIMIT $3;", slug, sinceVal, limitVal)
		} else if limit && since && desc {
			rows, err = db.Query("SELECT * FROM Users WHERE nickname IN (SELECT author FROM Threads WHERE forum=$1 AND author<$2 GROUP BY author) OR nickname IN (SELECT author FROM Posts WHERE forum=$1 AND author<$2 GROUP BY author) AND nickname<$2 ORDER BY nickname DESC LIMIT $3;", slug, sinceVal, limitVal)
		}

		if err != nil {
			// e := new(Error)
			var e models.Error
			e.Message = "Can't find forum with slug " + slug + "\n"
			resp, _ := json.Marshal(e)
			w.Header().Set("content-type", "application/json")

			w.WriteHeader(http.StatusNotFound)
			w.Write(resp)
			return
		}

		defer rows.Close()

		users := make([]models.User, 0)

		for rows.Next() {
			// usr := User{}
			var usr models.User

			err := rows.Scan(&usr.About, &usr.Email, &usr.Fullname, &usr.Nickname)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			users = append(users, usr)
		}

		resp, _ := json.Marshal(users)
		w.Header().Set("content-type", "application/json")

		w.Write(resp)

		return
	}
	return
}

func postDetails(w http.ResponseWriter, r *http.Request) { //GET + //POST +
	if r.Method == http.MethodGet {
		w.Header().Set("content-type", "application/json")

		var ID = mux.Vars(r)["id"]

		related := r.URL.Query().Get("related")
		relAdds := strings.Split(related, ",")

		// var postFull models.PostFull
		var post models.Post

		err := db.QueryRow("SELECT * FROM Posts WHERE id = $1;", ID).Scan(
			&post.Author,
			&post.Created,
			&post.Forum,
			ID,
			&post.IsEdited,
			&post.Message,
			&post.Parent,
			&post.Thread)

		if err != nil {
			if err == sql.ErrNoRows {
				var e models.Error
				e.Message = "Can't find user with id " + ID
				resData, _ := json.Marshal(e.Message)
				w.WriteHeader(http.StatusNotFound)
				w.Write(resData)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var postFull models.PostFull

		for _, data := range relAdds {
			if data == "user" {
				var postUser models.User
				row := db.QueryRow("SELECT * FROM User WHERE nickname = 1$;", post.Author)
				err := row.Scan(
					&postUser.About,
					&postUser.Email,
					&postUser.Fullname,
					&postUser.Nickname)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				postFull.Author = &postUser
			}

			if data == "thread" {
				var postThread models.Thread
				row := db.QueryRow("SELECT * FROM forum WHERE id = 1$;", post.Thread)
				err := row.Scan(
					&postThread.Author,
					&postThread.Created,
					&postThread.Forum,
					&postThread.Id,
					&postThread.Message,
					&postThread.Slug,
					&postThread.Title,
					&postThread.Votes)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				postFull.Thread = &postThread
			}
			if data == "forum" {
				var postForum models.Forum
				row := db.QueryRow("SELECT * FROM Forum WHERE slug = 1$;", post.Forum)
				err := row.Scan(
					&postForum.Posts,
					&postForum.Slug,
					&postForum.Threads,
					&postForum.Title,
					&postForum.User)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				postFull.Forum = &postForum
			}
		}

		resData, _ := json.Marshal(postFull)
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
		return
	}

	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		args := mux.Vars(r)
		ID := args["id"]

		var updatePost models.Post
		var oldPost models.Post
		var currentPost models.Post

		err = json.Unmarshal(reqBody, &updatePost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = db.QueryRow("SELECT * FROM Posts WHERE id = &1 ;", ID).Scan(
			&oldPost.Id,
			&oldPost.Parent,
			&oldPost.Author,
			&oldPost.Message,
			&oldPost.IsEdited,
			&oldPost.Forum,
			&oldPost.Thread,
			&oldPost.Created)
		if err != nil {
			if updatePost.Message != "" && oldPost.Message != updatePost.Message {
				err = db.QueryRow("UPDATE Posts SET message = $1, isedited = true WHERE id = $2 RETURNING *;",
					currentPost.Author,
					ID).Scan(
					&currentPost.Id,
					&currentPost.Author,
					&currentPost.Message,
					&currentPost.IsEdited,
					&currentPost.Forum,
					&currentPost.Thread,
					&currentPost.Created)
			} else {
				currentPost = oldPost
			}
		}

		if err != nil {
			if err == sql.ErrNoRows {
				var e models.Error

				tmp := int(currentPost.Author)
				pAuth := strconv.Itoa(tmp)

				e.Message = "Can't find user with id = " + pAuth + "\n"
				resData, _ := json.Marshal(e.Message)
				w.WriteHeader(http.StatusNotFound)
				w.Write(resData)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resData, _ := json.Marshal(currentPost)
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
		return
	}
}

func serviceClear(w http.ResponseWriter, r *http.Request) { //POST +
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")

		// sqlQuery := `
		// TRUNCATE TABLE post CASCADE;
		// TRUNCATE TABLE forumUser CASCADE;
		// TRUNCATE TABLE forum CASCADE;
		// TRUNCATE TABLE thread CASCADE;
		// TRUNCATE TABLE vote CASCADE;`

		_, err := db.Query("TRUNCATE TABLE Users, Forums, Threads, Posts, Votes")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
	return

}

func serviceStatus(w http.ResponseWriter, r *http.Request) { //GET +
	if r.Method == http.MethodGet {
		w.Header().Set("content-type", "application/json")
		var status models.Status

		err := db.QueryRow("SELECT COUNT(*) FROM Users;").Scan(&status.User)
		err = db.QueryRow("SELECT COUNT(*) FROM Forums;").Scan(&status.Forum)
		err = db.QueryRow("SELECT COUNT(*) FROM Threads;").Scan(&status.Thread)
		err = db.QueryRow("SELECT COUNT(*) FROM Posts;").Scan(&status.Post)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resData, _ := json.Marshal(status)
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
	}
	return
}

// func postsCreate(w http.ResponseWriter, r *http.Request) { //POST !-
// 	if r.Method == http.MethodPost {
// 		w.Header().Set("content-type", "application/json")

// 		reqBody, err := ioutil.ReadAll(r.Body)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		defer r.Body.Close()

// 		args := mux.Vars(r)
// 		slugOrId := args["slug_or_id"]

// 		// posts := make([]models.Post, 0)
// 		var newPosts []models.Post
// 		currentTime := time.Now()

// 		err = json.Unmarshal(reqBody, &newPosts)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		threadById, err := getThreadById(slugOrId)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				var e models.Error
// 				e.Message = "Can't find user with id = " + slugOrId + "\n"
// 				resData, _ := json.Marshal(e.Message)
// 				w.WriteHeader(http.StatusNotFound)
// 				w.Write(resData)
// 				return
// 			}
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		for index, posts := range newPosts {

// 			// if posts.Parent == 0 {
// 			// 	err := db.QueryRow("INSERT INTO Posts(author, forum, message, parent, thread) VALUES ($1,$2,$3,$4,$5) RETURNING id, created",
// 			// 		posts.Author,
// 			// 		thr.Forum,
// 			// 		posts.Message,
// 			// 		posts.Parent,
// 			// 		thr.Id).Scan(&id, &firstCreated)
// 			// }
// 		}

// 	}
// 	return
// }

func threadDetails(w http.ResponseWriter, r *http.Request) { //POST + //GET +
	if r.Method == http.MethodGet { // +
		w.Header().Set("content-type", "application/json")

		args := mux.Vars(r)
		slugOrId := args["slug_or_id"]

		thread, err := getThreadById(slugOrId)
		if err != nil {
			var e models.Error
			e.Message = "Can't find user with slug " + slugOrId
			resData, _ := json.Marshal(e)

			w.WriteHeader(http.StatusNotFound)
			w.Write(resData)
			return
		}

		resData, _ := json.Marshal(thread)
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
		return
	}

	if r.Method == http.MethodPost { // +
		w.Header().Set("content-type", "application/json")

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer r.Body.Close()

		args := mux.Vars(r)
		slugOrId := args["slug_or_id"]

		var returningThread models.Thread

		updateThread, err := getThreadById(slugOrId)
		if err != nil {
			var e models.Error
			e.Message = "Can't find user with slug " + slugOrId
			resData, _ := json.Marshal(e)

			w.WriteHeader(http.StatusNotFound)
			w.Write(resData)
			return
		}

		err = json.Unmarshal(reqBody, &updateThread)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		threadSlugOrId, err := strconv.Atoi(slugOrId)
		var adds string
		if err != nil {
			adds = "slug='" + slugOrId + "' "
		} else {
			adds = "id=" + strconv.Itoa(threadSlugOrId)
		}

		if updateThread.Message == "" && updateThread.Title == "" {
			err := db.QueryRow("SELECT * FROM Threads WHERE "+adds+";").Scan(
				&returningThread.Author,
				&returningThread.Created,
				&returningThread.Forum,
				&returningThread.Id,
				&returningThread.Message,
				&returningThread.Slug,
				&returningThread.Title,
				&returningThread.Votes)
			if err != nil {
				if err == sql.ErrNoRows {
					var e models.Error
					e.Message = "Can't find thread with id or slug " + slugOrId
					resData, _ := json.Marshal(e.Message)
					w.WriteHeader(http.StatusNotFound)
					w.Write(resData)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if updateThread.Message != "" && updateThread.Title != "" {
			err := db.QueryRow("UPDATE Threads SET message = $1, title = $2 WHERE "+adds+" RETURNING *;",
				&updateThread.Message,
				&updateThread.Title).Scan(
				&returningThread.Author,
				&returningThread.Created,
				&returningThread.Forum,
				&returningThread.Id,
				&returningThread.Message,
				&returningThread.Slug,
				&returningThread.Title,
				&returningThread.Votes)
			if err != nil {
				if err == sql.ErrNoRows {
					var e models.Error
					e.Message = "Can't find thread with id or slug " + slugOrId
					resData, _ := json.Marshal(e.Message)
					w.WriteHeader(http.StatusNotFound)
					w.Write(resData)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		} else if updateThread.Message == "" && updateThread.Title != "" {
			err := db.QueryRow("UPDATE Threads SET title = $1 WHERE "+adds+" RETURNING *;",
				&updateThread.Title).Scan(
				&returningThread.Author,
				&returningThread.Created,
				&returningThread.Forum,
				&returningThread.Id,
				&returningThread.Message,
				&returningThread.Slug,
				&returningThread.Title,
				&returningThread.Votes)
			if err != nil {
				if err == sql.ErrNoRows {
					var e models.Error
					e.Message = "Can't find thread with id or slug " + slugOrId
					resData, _ := json.Marshal(e.Message)
					w.WriteHeader(http.StatusNotFound)
					w.Write(resData)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if updateThread.Message != "" && updateThread.Title == "" {
			err := db.QueryRow("UPDATE Threads SET message = $1 WHERE "+adds+" RETURNING *;",
				&updateThread.Message).Scan(
				&returningThread.Author,
				&returningThread.Created,
				&returningThread.Forum,
				&returningThread.Id,
				&returningThread.Message,
				&returningThread.Slug,
				&returningThread.Title,
				&returningThread.Votes)
			if err != nil {
				if err == sql.ErrNoRows {
					var e models.Error
					e.Message = "Can't find thread with id or slug " + slugOrId
					resData, _ := json.Marshal(e.Message)
					w.WriteHeader(http.StatusNotFound)
					w.Write(resData)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resData, _ := json.Marshal(returningThread)
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
		return
	}
	return
}

func threadPosts(w http.ResponseWriter, r *http.Request) { //GET - (Sort) СложнАААААААААА

}

func threadVote(w http.ResponseWriter, r *http.Request) { //POST +
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		args := mux.Vars(r)
		slugOrID := args["slug_or_id"]

		// var returningThread models.Thread
		var newUserVote models.Vote
		var oldUserVote models.Vote

		returningThread, err := getThreadById(slugOrID)
		if err != nil {
			var e models.Error
			e.Message = "Can't find user with slug " + slugOrID
			resData, _ := json.Marshal(e)

			w.WriteHeader(http.StatusNotFound)
			w.Write(resData)
			return
		}

		err = json.Unmarshal(reqBody, &newUserVote)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		threadSlugOrId, err := strconv.Atoi(slugOrID)
		var adds string
		if err != nil {
			adds = "slug='" + slugOrID + "' "
		} else {
			adds = "id=" + strconv.Itoa(threadSlugOrId)
		}

		row := db.QueryRow("SELECT voice FROM Votes WHERE nickname=$1", oldUserVote.Nickname)
		err = row.Scan(&oldUserVote.Voice)
		if err != nil {
			_, err = db.Exec("INSERT INTO Votes(nickname, voice) VALUES ($1,$2) ", &newUserVote.Nickname, newUserVote.Voice)
			err = db.QueryRow("UPDATE Threads SET nickname = $1, votes = votes+$1 WHERE "+adds+" RETURNING *;",
				&newUserVote.Nickname,
				&newUserVote.Voice).Scan(
				&returningThread.Author,
				&returningThread.Created,
				&returningThread.Forum,
				&returningThread.Id,
				&returningThread.Message,
				&returningThread.Slug,
				&returningThread.Title,
				&returningThread.Votes)

			resData, _ := json.Marshal(returningThread)
			w.WriteHeader(http.StatusOK)
			w.Write(resData)
			return

		} else {
			row.Scan(&oldUserVote.Voice)
			if oldUserVote.Voice != newUserVote.Voice {
				if oldUserVote.Voice == -1 {
					_, err = db.Exec("UPDATE Threads SET votes=votes-2 WHERE " + adds + ";")
				} else {
					_, err = db.Exec("UPDATE threads SET votes=votes+2 WHERE " + adds + ";")
				}
			}

			err := db.QueryRow("SELECT * FROM Threads WHERE "+adds+";").
				Scan(&returningThread.Author,
					&returningThread.Created,
					&returningThread.Forum,
					&returningThread.Id,
					&returningThread.Message,
					&returningThread.Slug,
					&returningThread.Title,
					&returningThread.Votes)
			if err != nil {
				if err == sql.ErrNoRows {
					var e models.Error
					e.Message = "Can't find thread author by slug or id" + slugOrID
					resData, _ := json.Marshal(e.Message)
					w.WriteHeader(http.StatusNotFound)
					w.Write(resData)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resData, _ := json.Marshal(returningThread)
			w.WriteHeader(http.StatusOK)
			w.Write(resData)
			return
		}
	}
	return
}

func userCreate(w http.ResponseWriter, r *http.Request) { //POST +
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		reqBody, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// user := new(models.User)
		var user models.User
		// var user models.User
		var newUser models.User

		err = json.Unmarshal(reqBody, &newUser)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		args := mux.Vars(r)
		nickname := args["nickname"]
		newUser.Nickname = nickname

		err = db.QueryRow("INSERT INTO Users (about, email, fullname, nickname) VALUES ($1, $2, $3, $4) RETURNING *",
			newUser.About,
			newUser.Email,
			newUser.Fullname,
			newUser.Nickname).
			Scan(
				&user.About,
				&user.Email,
				&user.Fullname,
				&user.Nickname)

		if err != nil {
			// fmt.Println(err.Error)

			var existUser []models.User

			row, err := db.Query("SELECT * FROM Users WHERE nickname=$1 OR email=$2", newUser.Nickname, newUser.Email)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			for row.Next() {
				var rowUser models.User
				err := row.Scan(&rowUser.About, &rowUser.Email, &rowUser.Fullname, &rowUser.Nickname)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				existUser = append(existUser, rowUser)
			}

			resData, _ := json.Marshal(existUser)
			w.WriteHeader(http.StatusConflict)
			w.Write(resData)
			return
		}

		resData, _ := json.Marshal(newUser)
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
		return
	}
	return
}

func userProfile(w http.ResponseWriter, r *http.Request) { //GET + //POST +
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		reqBody, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var oldUser models.User
		var newUser models.User
		var userUpdate models.UserUpdate

		err = json.Unmarshal(reqBody, &userUpdate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// fmt.Println(userUpdate)
		// fmt.Println("-----------------------")

		args := mux.Vars(r)
		nickname := args["nickname"]
		// userUpdate.Nickname = nickname
		err = db.QueryRow("SELECT * FROM users WHERE nickname=$1", nickname).
			Scan(
				&oldUser.About,
				&oldUser.Email,
				&oldUser.Fullname,
				&oldUser.Nickname)

		if err != nil {
			if err == sql.ErrNoRows {
				var e models.Error
				e.Message = "Can't find user with nickname " + nickname
				resData, _ := json.Marshal(e)

				w.WriteHeader(http.StatusNotFound)
				w.Write(resData)
				return
			}
		}

		if userUpdate.Email == "" {
			userUpdate.Email = oldUser.Email
		}
		if userUpdate.Fullname == "" {
			userUpdate.Fullname = oldUser.Fullname
		}
		if userUpdate.About == "" {
			userUpdate.About = oldUser.About
		}

		err = db.QueryRow("UPDATE Users SET about=$1, email=$2, fullname=$3 WHERE nickname=$4 RETURNING *;",
			userUpdate.About,
			userUpdate.Email,
			userUpdate.Fullname,
			nickname).Scan(
			&newUser.About,
			&newUser.Email,
			&newUser.Fullname,
			&newUser.Nickname)

		// fmt.Println("-----------------------")
		// fmt.Println(newUser)
		// fmt.Println("-----------------------")

		if err != nil {
			var e models.Error
			e.Message = "Can't change prifile with id " + nickname
			resData, _ := json.Marshal(e)

			w.WriteHeader(http.StatusConflict)
			w.Write(resData)
			return
		}

		resData, _ := json.Marshal(newUser)
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("content-type", "application/json")

		args := mux.Vars(r)
		nickname := args["nickname"]
		var user models.User
		user.Nickname = nickname

		err := db.QueryRow("SELECT * FROM Users WHERE nickname = $1", &user.Nickname).
			Scan(&user.About,
				&user.Email,
				&user.Fullname,
				&user.Nickname)
		if err != nil {
			if err == sql.ErrNoRows {
				var e models.Error
				e.Message = "Can't find user with nickname " + nickname
				resData, _ := json.Marshal(e)

				w.WriteHeader(http.StatusNotFound)
				w.Write(resData)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resData, _ := json.Marshal(user)
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
		return
	}
	return
}

func getThreadById(slug string) (*models.Thread, error) {
	threadID, err := strconv.Atoi(slug)
	var row *sql.Row
	if err != nil {
		row = db.QueryRow("SELECT * FROM Threads WHERE slug=$1;", slug)
	} else {
		row = db.QueryRow("SELECT * FROM Threads WHERE id=$1;", threadID)
	}

	thread := new(models.Thread)
	err = row.Scan(&thread.Id,
		&thread.Author,
		&thread.Created,
		&thread.Forum,
		&thread.Message,
		&thread.Slug,
		&thread.Title,
		&thread.Votes)

	if err != nil {
		return nil, err
	}

	return thread, nil
}

func getForum(slugOrId string) *models.Forum {
	// forum := Forum{}
	var forum models.Forum
	err := db.QueryRow("SELECT * FROM forums WHERE slug=$1", slugOrId).Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)

	if err != nil {
		return nil
	}

	return &forum
}
