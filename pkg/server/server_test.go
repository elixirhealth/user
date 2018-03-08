package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser_ok(t *testing.T) {
	config := NewDefaultConfig()
	c, err := newUser(config)
	assert.Nil(t, err)
	assert.Equal(t, config, c.config)
	// TODO assert.NotEmpty on other elements of server struct
	//assert.NotEmpty(t, c.storer)
}

func TestNewUser_err(t *testing.T) {
	badConfigs := map[string]*Config{
	// TODO add bad config instances
	}
	for desc, badConfig := range badConfigs {
		c, err := newUser(badConfig)
		assert.NotNil(t, err, desc)
		assert.Nil(t, c)
	}
}

// TODO add TestUser_ENDPOINT_(ok|err) for each ENDPOINT
