package command

import "errors"

var ErrEmptyHunks = errors.New("modify file must contain at least one hunk")
