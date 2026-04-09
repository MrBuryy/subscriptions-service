package service

import "errors"

var (
	ErrInvalidID        = errors.New("invalid id")
	ErrInvalidUserID    = errors.New("invalid user_id")
	ErrInvalidPrice     = errors.New("invalid price")
	ErrInvalidService   = errors.New("invalid service_name")
	ErrInvalidStartDate = errors.New("invalid start_date")
	ErrInvalidEndDate   = errors.New("invalid end_date")
	ErrNotFound         = errors.New("not found")
)