package alicloud

import (
	"testing"
)

func TestContainsCidr(t *testing.T) {

	if contains, err := ContainsCidr("0.0.0.0/0", "192.168.3.0/24"); err != nil || !contains {
		t.Logf("fail on test default cidr contains node cidr")
		t.Fail()
	}

	if contains, err := ContainsCidr("192.168.0.0/16", "192.168.3.0/24"); err != nil || !contains {
		t.Logf("fail on test cluster cidr contains node cidr")
		t.Fail()
	}

	if contains, err := ContainsCidr("192.168.3.0/24", "192.168.0.0/16"); err != nil || contains {
		t.Logf("fail on test node cidr not contains cluster cidr")
		t.Fail()
	}

	if _, err := ContainsCidr("", "192.168.0.0/16"); err == nil {
		t.Logf("fail on test node cidr not contains cluster cidr")
		t.Fail()
	}
}

func TestRealContainsCidr(t *testing.T) {

	if contains, err := RealContainsCidr("0.0.0.0/0", "192.168.3.0/24"); err != nil || !contains {
		t.Logf("fail on test default cidr real contains node cidr")
		t.Fail()
	}

	if contains, err := RealContainsCidr("192.168.0.0/16", "192.168.3.0/24"); err != nil || !contains {
		t.Logf("fail on test cluster cidr real contains node cidr")
		t.Fail()
	}

	if contains, err := RealContainsCidr("192.168.3.0/24", "192.168.0.0/16"); err != nil || contains {
		t.Logf("fail on test node cidr not real contains cluster cidr")
		t.Fail()
	}

	if contains, err := RealContainsCidr("192.168.3.0/24", "192.168.3.0/24"); err != nil || contains {
		t.Logf("fail on test node cidr not real contains same route cidr")
		t.Fail()
	}

	if _, err := RealContainsCidr("", "192.168.0.0/16"); err == nil {
		t.Logf("fail on test node cidr not real contains cluster cidr")
		t.Fail()
	}
}