package storageErrors

import "errors"

var (
	//ErrInternalError  = errors.New("internal error")
	ErrDuplicateEntry = errors.New("duplicate entry")
)
