package ytapi

import (
	"testing"

	"github.com/lbryio/lbry.go/v2/extras/stop"
	"github.com/lbryio/ytsync/v5/ip_manager"
	"github.com/stretchr/testify/assert"
)

func TestChannelInfo(t *testing.T) {
	ipPool, err := ip_manager.GetIPPool(stop.New())
	assert.NoError(t, err)
	info, err := ChannelInfo("UCNQfQvFMPnInwsU_iGYArJQ", 0, ipPool)
	assert.NoError(t, err)
	assert.NotNil(t, info)
}
