package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

func (s Service) Create(task Task) (Task, error) {
	id, err := s.store.Add(task)
	if err != nil {
		return task, err
	}

	task.Id = fmt.Sprint(id)
	return task, nil
}

func (s Service) FindBy(search string) ([]Task, error) {
	if search == "" {
		return s.store.GetAll()
	}
	if d, err := time.Parse("02.01.2006", search); err == nil {
		return s.store.GetByDate(d)
	}
	return s.store.GetByTitle(search)
}

func (s Service) FindById(id int) (Task, error) {
	return s.store.GetById(id)
}

func (s Service) Update(task Task) error {
	id, err := strconv.Atoi(task.Id)
	if err != nil {
		return err
	}

	if _, err := s.FindById(id); err != nil {
		return err
	}

	return s.store.Update(task)
}

func (s Service) Delete(id int) error {
	if _, err := s.FindById(id); err != nil {
		return err
	}

	return s.store.Delete(id)
}

func (s Service) NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("не указаны параметры")
	}

	startDate, err := time.Parse(DATE_FORMAT, date)
	if err != nil {
		return "", err
	}

	parts := strings.Split(repeat, " ")
	param := parts[0]
	if param == "y" {
		return s.calculateNextYearDate(now, startDate)
	}
	if param == "d" {
		if len(parts) < 2 {
			return "", fmt.Errorf("не указано число дней в интервале")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", err
		}
		if days > 400 {
			return "", fmt.Errorf("превышен максимально допустимый интервал")
		}
		return s.calculateNextDayDate(now, startDate, days)
	}
	return "", fmt.Errorf("неподдерживаемый формат %s", param)
}

func (s Service) calculateNextYearDate(now, startDate time.Time) (string, error) {
	currDate := startDate.AddDate(1, 0, 0)
	for now.After(currDate) || now.Equal(currDate) {
		currDate = currDate.AddDate(1, 0, 0)
	}
	return currDate.Format(DATE_FORMAT), nil
}

func (s Service) calculateNextDayDate(now, startDate time.Time, days int) (string, error) {
	currDate := startDate.AddDate(0, 0, days)
	for now.After(currDate) {
		currDate = currDate.AddDate(0, 0, days)
	}
	return currDate.Format(DATE_FORMAT), nil
}
