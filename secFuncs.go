package main

import (
	"database/sql"
	"strconv"

	models "./models"
)

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
