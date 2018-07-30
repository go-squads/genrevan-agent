package config_test

import (
	"testing"

	"github.com/go-squads/genrevan-agent/config"
	"github.com/stretchr/testify/assert"
)

func TestConfig_ExpectedNoErrorWhileReadConfig(t *testing.T) {
	err := config.SetupConfig()

	assert.Equal(t, nil, err)
}
