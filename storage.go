package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return Store{db: db}
}

const limitTask = 50

func (s Store) Add(task Task) (int, error) {
	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := s.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s Store) GetAll() ([]Task, error) {
	return s.queryTasks("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?", limitTask)
}

func (s Store) GetByDate(date time.Time) ([]Task, error) {
	return s.queryTasks("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? LIMIT ?", date.Format(DATE_FORMAT), limitTask)
}

func (s Store) GetById(id int) (Task, error) {
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	return s.queryTask(query, id)
}

func (s Store) GetByTitle(search string) ([]Task, error) {
	searchPattern := fmt.Sprintf("%%%s%%", strings.ToUpper(search))
	return s.queryTasks("SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR UPPER(comment) LIKE ? ORDER BY date LIMIT ?", searchPattern, searchPattern, limitTask)
}

func (s Store) Update(task Task) error {
	query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
	_, err := s.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.Id)
	return err
}

func (s Store) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	return err
}

func (s Store) queryTasks(query string, args ...interface{}) ([]Task, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s Store) queryTask(query string, args ...interface{}) (Task, error) {
	var task Task
	row := s.db.QueryRow(query, args...)
	if err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		return task, err
	}
	return task, nil
}
