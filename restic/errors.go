package restic

import "errors"

var (
	ErrRepoLocked       = errors.New("repo already locked")
	ErrConnectionFailed = errors.New("connection failed")
	ErrCheckFailed      = errors.New("integrity check failed")
	ErrUnknown          = errors.New("unknown error")
)
