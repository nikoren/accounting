package domain

import "errors"

// SplitStatus represents whether a split is still editable or has been finalized.
type SplitStatus string

const (
	SplitStatusDraft     SplitStatus = "draft"
	SplitStatusFinalized SplitStatus = "finalized"
)

// ErrNotFound is returned when a requested resource is not found
var ErrNotFound = errors.New("not found")
