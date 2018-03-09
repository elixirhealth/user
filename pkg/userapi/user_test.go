package userapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAddEntityRequest(t *testing.T) {
	userID, entityID := "some user ID", "some entity ID"
	cases := map[string]struct {
		rq       *AddEntityRequest
		expected error
	}{
		"ok": {
			rq: &AddEntityRequest{
				UserId:   userID,
				EntityId: entityID,
			},
			expected: nil,
		},
		"empty entity ID": {
			rq: &AddEntityRequest{
				UserId: userID,
			},
			expected: ErrEmptyEntityID,
		},
		"empty user ID": {
			rq: &AddEntityRequest{
				EntityId: entityID,
			},
			expected: ErrEmptyUserID,
		},
	}
	for desc, c := range cases {
		err := ValidateAddEntityRequest(c.rq)
		assert.Equal(t, c.expected, err, desc)
	}
}

func TestValidateGetEntitiesRequest(t *testing.T) {
	userID := "some user ID"
	cases := map[string]struct {
		rq       *GetEntitiesRequest
		expected error
	}{
		"ok": {
			rq: &GetEntitiesRequest{
				UserId: userID,
			},
			expected: nil,
		},
		"empty user ID": {
			rq:       &GetEntitiesRequest{},
			expected: ErrEmptyUserID,
		},
	}
	for desc, c := range cases {
		err := ValidateGetEntitiesRequest(c.rq)
		assert.Equal(t, c.expected, err, desc)
	}
}
