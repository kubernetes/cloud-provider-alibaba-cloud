package base

import (
	"os"
	"testing"

	aliCredentials "github.com/aliyun/credentials-go/credentials"
	"github.com/stretchr/testify/assert"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
)

func setRRSAEnvVars(t *testing.T) {
	t.Helper()
	t.Setenv(aliCredentials.ENVOIDCProviderArn, "acs:ram::1234567890:oidc-provider/ack-rrsa-test")
	t.Setenv(aliCredentials.ENVOIDCTokenFile, "/var/run/secrets/ack-rrsa/token")
	t.Setenv(aliCredentials.ENVRoleArn, "acs:ram::1234567890:role/test-role")
}

func clearAKEnvVars(t *testing.T) {
	t.Helper()
	t.Setenv(AccessKeyID, "")
	t.Setenv(AccessKeySecret, "")
}

func TestIsRRSAEnabled_AllSet(t *testing.T) {
	setRRSAEnvVars(t)
	assert.True(t, isRRSAEnabled())
}

func TestIsRRSAEnabled_NoneSet(t *testing.T) {
	t.Setenv(aliCredentials.ENVOIDCProviderArn, "")
	t.Setenv(aliCredentials.ENVOIDCTokenFile, "")
	t.Setenv(aliCredentials.ENVRoleArn, "")
	assert.False(t, isRRSAEnabled())
}

func TestIsRRSAEnabled_PartialSet(t *testing.T) {
	t.Setenv(aliCredentials.ENVOIDCProviderArn, "acs:ram::123:oidc-provider/test")
	t.Setenv(aliCredentials.ENVOIDCTokenFile, "")
	t.Setenv(aliCredentials.ENVRoleArn, "acs:ram::123:role/test")
	assert.False(t, isRRSAEnabled())
}

func TestGetTokenAuth_RRSAPriority(t *testing.T) {
	// Clear AK env vars and cloud config
	clearAKEnvVars(t)
	origAK := ctrlCfg.CloudCFG.Global.AccessKeyID
	origSK := ctrlCfg.CloudCFG.Global.AccessKeySecret
	ctrlCfg.CloudCFG.Global.AccessKeyID = ""
	ctrlCfg.CloudCFG.Global.AccessKeySecret = ""
	defer func() {
		ctrlCfg.CloudCFG.Global.AccessKeyID = origAK
		ctrlCfg.CloudCFG.Global.AccessKeySecret = origSK
	}()

	// Ensure addon token file does not exist (it shouldn't in test env)
	_, err := os.Stat(AddonTokenFilePath)
	if err == nil {
		t.Skip("addon token file exists, skipping")
	}

	setRRSAEnvVars(t)

	mgr := &ClientMgr{Region: "cn-hangzhou"}
	tokenAuth := mgr.GetTokenAuth()
	_, ok := tokenAuth.(*RRSAToken)
	assert.True(t, ok, "expected RRSAToken, got %T", tokenAuth)
}

func TestGetTokenAuth_AKOverRRSA(t *testing.T) {
	// Set both AK and RRSA env vars — AK should win
	t.Setenv(AccessKeyID, "test-ak")
	t.Setenv(AccessKeySecret, "test-sk")
	setRRSAEnvVars(t)

	origAK := ctrlCfg.CloudCFG.Global.AccessKeyID
	origSK := ctrlCfg.CloudCFG.Global.AccessKeySecret
	ctrlCfg.CloudCFG.Global.AccessKeyID = ""
	ctrlCfg.CloudCFG.Global.AccessKeySecret = ""
	defer func() {
		ctrlCfg.CloudCFG.Global.AccessKeyID = origAK
		ctrlCfg.CloudCFG.Global.AccessKeySecret = origSK
	}()

	mgr := &ClientMgr{Region: "cn-hangzhou"}
	tokenAuth := mgr.GetTokenAuth()
	_, ok := tokenAuth.(*AkAuthToken)
	assert.True(t, ok, "expected AkAuthToken when AK env vars are set, got %T", tokenAuth)
}
