package models

import (
	"errors"
	"time"
)

// ошибка поиска записи в бд
var ErrNoRecord = errors.New("models: подходящей записи не найдено")

// структура заметки в бд
type Snippets struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}
