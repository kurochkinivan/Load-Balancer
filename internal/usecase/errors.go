package usecase

import "errors"

var (
	ErrClientNotFound = errors.New("client was no found")
	ErrClientExists   = errors.New("client already exists")
)
