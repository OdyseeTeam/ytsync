package ytapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChannelInfo(t *testing.T) {
	info, err := ChannelInfo("UCNQfQvFMPnInwsU_iGYArJQ", 0)
	assert.NoError(t, err)
	assert.NotNil(t, info)
}
