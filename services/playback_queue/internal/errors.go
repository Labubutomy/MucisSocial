package internal

import "errors"

var (
	ErrBadRequest    = errors.New("bad request")
	ErrEmptyQueue    = errors.New("queue is empty")
	ErrQueueNotFound = errors.New("queue not found")
)
