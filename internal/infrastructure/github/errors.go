package github

import "errors"

var (
	ErrNotFound             = errors.New("file not found")
	ErrMultipleMatches      = errors.New("multiple matches found")
	ErrSearchStringNotFound = errors.New("search string not found")
	ErrApplyHunksFailed     = errors.New("failed to apply hunks")
)
