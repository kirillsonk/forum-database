package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	models "./models"
	"github.com/gorilla/mux"
)

func createForum(w http.ResponseWriter, r *http.Request) { //POST +
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var forum models.Forum

		err = json.Unmarshal(reqBody, &forum)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO Forums (slug, title, author) VALUES ($1 , $2, $3)",
			&forum.Slug,
			&forum.Title,
			&forum.User)
		if err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"forums_author_key\"" {
				row := db.QueryRow("SELECT * FROM forums WHERE author=$1", forum.User)
				row.Scan(&forum.Posts,
					&forum.Slug,
					&forum.Threads,
					&forum.Title,
					&forum.User)

				resData, _ := json.Marshal(forum)
				w.WriteHeader(http.StatusConflict)
				w.Write(resData)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = db.QueryRow("SELECT * FROM Forums WHERE \"author\" = $1;", forum).Scan(&forum.User)
		if err != nil {
			if err == sql.ErrNoRows {
				var e models.Error
				e.Message = "Can't find user with id " + forum.User
				resData, _ := json.Marshal(e.Message)
				w.WriteHeader(http.StatusNotFound)
				w.Write(resData)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resData, _ := json.Marshal(forum)
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
		return
	}

	return
}

func createThread(w http.ResponseWriter, r *http.Request) { //POST +
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer r.Body.Close()

		var thread models.Thread

		err = json.Unmarshal(reqBody, &thread)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		args := mux.Vars(r)
		Slug := args["slug"]

		_, err = db.Exec("INSERT INTO Threads (author, created, message, title, slug) VALUES ($1 , $2, $3, $4, $5)",
			&thread.Author,
			&thread.Created,
			&thread.Message,
			&thread.Title,
			Slug)
		if err.Error() == "pq: duplicate key value violates unique constraint \"threads_title_key\"" {
			row := db.QueryRow("SELECT * FROM forums WHERE title=$1", thread.Title)
			row.Scan(&thread.Author,
				&thread.Created,
				&thread.Forum,
				&thread.Id,
				&thread.Message,
				&thread.Slug,
				&thread.Title,
				&thread.Votes,
			)

			resData, _ := json.Marshal(thread)
			w.WriteHeader(http.StatusConflict)
			w.Write(resData)
			return
		}

		err = db.QueryRow("SELECT * FROM Threads WHERE \"author\" = $1 OR \"forum\" = $2;", thread).Scan(
			&thread.Author,
			&thread.Forum)
		if err != nil {
			if err == sql.ErrNoRows {
				var e models.Error
				e.Message = "Can't find user with id " + thread.Author
				resData, _ := json.Marshal(e.Message)
				w.WriteHeader(http.StatusNotFound)
				w.Write(resData)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resData, _ := json.Marshal(thread)
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
		return
	}
	return
}

func forumDetails(w http.ResponseWriter, r *http.Request) { //GET +
	if r.Method == http.MethodGet {
		w.Header().Set("content-type", "application/json")

		args := mux.Vars(r)
		Slug := args["slug"]

		getDataBySlug := db.QueryRow("SELECT * FROM Forums WHERE slug = $1;", Slug)

		var forum models.Forum

		err := getDataBySlug.Scan(&forum.Posts,
			Slug,
			&forum.Threads,
			&forum.Title,
			&forum.User)
		if err != nil {
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

func forumThreads(w http.ResponseWriter, r *http.Request) { //GET (Sort) -
	if r.Method == http.MethodGet {
		return
	}
	return
}

func forumUsers(w http.ResponseWriter, r *http.Request) { //GET (Sort) -
	if r.Method == http.MethodGet {
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

func postsCreate(w http.ResponseWriter, r *http.Request) { //POST !-
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		args := mux.Vars(r)
		slugOrId := args["slug_or_id"]

		// posts := make([]models.Post, 0)
		var newPosts []models.Post
		currentTime := time.Now()

		err = json.Unmarshal(reqBody, &newPosts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		threadById, err := getThreadById(slugOrId)
		if err != nil {
			if err == sql.ErrNoRows {
				var e models.Error
				e.Message = "Can't find user with id = " + slugOrId + "\n"
				resData, _ := json.Marshal(e.Message)
				w.WriteHeader(http.StatusNotFound)
				w.Write(resData)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for index, posts := range newPosts {

			// if posts.Parent == 0 {
			// 	err := db.QueryRow("INSERT INTO Posts(author, forum, message, parent, thread) VALUES ($1,$2,$3,$4,$5) RETURNING id, created",
			// 		posts.Author,
			// 		thr.Forum,
			// 		posts.Message,
			// 		posts.Parent,
			// 		thr.Id).Scan(&id, &firstCreated)
			// }
		}

	}
	return
}

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

func threadPosts(w http.ResponseWriter, r *http.Request) { //GET (Sort) -

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
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer r.Body.Close()

		var user models.User
		var newUser models.User

		err = json.Unmarshal(reqBody, &newUser)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		args := mux.Vars(r)
		nickname := args["nickname"]
		user.Nickname = nickname

		err = db.QueryRow("INSERT INTO Users (about, email, fullname, nickname) VALUES ($1 , $2, $3, $4) RETURNING *",
			&user.About,
			&user.Email,
			&user.Fullname,
			&user.Nickname).
			Scan(
				&newUser.About,
				&newUser.Email,
				&newUser.Fullname,
				&newUser.Nickname)
			// if err.Error() == "pq: duplicate key value violates unique constraint \"Users_nickname_key\"" {
		if err != nil {

			var existUser []models.User

			row, err := db.Query("SELECT * FROM Users WHERE nickname=$1 OR email=$2")
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
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer r.Body.Close()

		var user models.User
		var userUpdate models.UserUpdate

		err = json.Unmarshal(reqBody, &userUpdate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		args := mux.Vars(r)
		nickname := args["nickname"]
		user.Nickname = nickname
		// UPDATE Posts SET message = $1, isedited = true WHERE id = $2 RETURNING *;
		err = db.QueryRow("UPDATE Users SET (about, email, fullname) VALUES ($1 , $2, $3) WHERE nickname=$4 RETURNING *",
			&userUpdate.About,
			&userUpdate.Email,
			&userUpdate.Fullname,
			nickname).
			Scan(
				&user.About,
				&user.Email,
				&user.Fullname,
				&user.Nickname)
		if err != nil { //404 +
			if err == sql.ErrNoRows {
				var e models.Error
				e.Message = "Can't find user with nickname " + nickname
				resData, _ := json.Marshal(e)

				w.WriteHeader(http.StatusNotFound)
				w.Write(resData)
				return
			} else { //409 +
				var e models.Error
				e.Message = "Can't find user with nickname " + nickname
				resData, _ := json.Marshal(e)

				w.WriteHeader(http.StatusConflict)
				w.Write(resData)
				return
			}
		}

		resData, _ := json.Marshal(user)
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
