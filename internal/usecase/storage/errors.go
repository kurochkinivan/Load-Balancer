package storage

import "errors"

var (
	ErrClientNotFound = errors.New("client was not found")
	ErrClientExists   = errors.New("client already exists")
)
