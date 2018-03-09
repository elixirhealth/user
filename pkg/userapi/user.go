package userapi

import "github.com/pkg/errors"

var (
	// ErrEmptyUserID denotes when the user ID field is empty.
	ErrEmptyUserID = errors.New("empty user ID field")

	// ErrEmptyEntityID denotes when the entity ID field is empty.
	ErrEmptyEntityID = errors.New("empty entity ID field")
)

// TODO add ValidateENDPOINTRequest method for each service ENDPOINT
