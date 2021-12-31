package base

import (
	"github.com/stretchr/testify/assert"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"testing"
)

func TestVswitchID(t *testing.T) {
	raw := ctrlCfg.CloudCFG.Global.VswitchID
	cfg := &CfgMetaData{base: NewBaseMetaData(nil)}

	ctrlCfg.CloudCFG.Global.VswitchID = "vsw-a"
	vsw, err := cfg.VswitchID()
	assert.Equal(t, nil, err)
	assert.Equal(t, "vsw-a", vsw)

	ctrlCfg.CloudCFG.Global.VswitchID = ":vsw-b"
	vsw, err = cfg.VswitchID()
	assert.Equal(t, nil, err)
	assert.Equal(t, "vsw-b", vsw)

	ctrlCfg.CloudCFG.Global.VswitchID = ":vsw-c,:vsw-d"
	vsw, err = cfg.VswitchID()
	assert.Equal(t, nil, err)
	assert.Equal(t, "vsw-c", vsw)

	ctrlCfg.CloudCFG.Global.VswitchID = "cn-hangzhou-h:vsw-h"
	vsw, err = cfg.VswitchID()
	assert.Equal(t, nil, err)
	assert.Equal(t, "vsw-h", vsw)

	ctrlCfg.CloudCFG.Global.VswitchID = "cn-hangzhou-h:vsw-e,cn-hangzhou-f:vsw-f"
	vsw, err = cfg.VswitchID()
	assert.Equal(t, nil, err)
	assert.Equal(t, "vsw-e", vsw)

	ctrlCfg.CloudCFG.Global.VswitchID = raw
}
