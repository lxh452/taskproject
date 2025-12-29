package task

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/mon"
)

var (
	ErrNotMongoFound   = mon.ErrNotFound
	ErrInvalidObjectId = errors.New("invalid objectId")
)
