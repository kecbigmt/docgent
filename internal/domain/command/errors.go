package command

import "errors"

var ErrEmptyModifyHunks = errors.New("modify file must contain at least one hunk")
