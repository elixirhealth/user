package userapi

import "github.com/pkg/errors"

var (
	// ErrEmptyUserID denotes when the user ID field is empty.
	ErrEmptyUserID = errors.New("empty user ID field")

	// ErrEmptyEntityID denotes when the entity ID field is empty.
	ErrEmptyEntityID = errors.New("empty entity ID field")
)

// ValidateAddEntityRequest checks that the entity and user ID fields are populated.
func ValidateAddEntityRequest(rq *AddEntityRequest) error {
	if rq.EntityId == "" {
		return ErrEmptyEntityID
	}
	if rq.UserId == "" {
		return ErrEmptyUserID
	}
	return nil
}

// ValidateGetEntitiesRequest checks that the user ID fields are populated.
func ValidateGetEntitiesRequest(rq *GetEntitiesRequest) error {
	if rq.UserId == "" {
		return ErrEmptyUserID
	}
	return nil
}
