package warp_test

import (
	"encoding/json"
	"testing"

	"github.com/cloudogu/ces-confd/confd/warp"
)

func TestTarget_MarshalJSON(t *testing.T) {
	testMarshalJSON(t, warp.TARGET_EXTERNAL, "{\"Target\":\"external\"}")
	testMarshalJSON(t, warp.TARGET_SELF, "{\"Target\":\"self\"}")

	if _, err := json.Marshal(&targetStruct{12}); err == nil {
		t.Errorf("marshal should fail because of an invalid value")
	}
}

func testMarshalJSON(t *testing.T, target warp.Target, expected string) {
	value := marshal(t, target)
	if value != expected {
		t.Errorf("value %s is not the expected %s", value, expected)
	}
}

func marshal(t *testing.T, target warp.Target) string {
	test := targetStruct{target}
	json, err := json.Marshal(&test)
	if err != nil {
		t.Errorf("failed to marshal test struct: %v", err)
	}
	return string(json)
}

type targetStruct struct {
	Target warp.Target
}