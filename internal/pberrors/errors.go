// Contains any errors used in the internals
package pberrors

import "errors"

var (
	ErrInvalidAgentID = errors.New("invalid agent ID")
	ErrInvalidTaskID  = errors.New("invalid task ID")
)
