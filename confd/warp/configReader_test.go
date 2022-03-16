package warp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigReader_readSupport(t *testing.T) {
	reader := &ConfigReader{
		configuration: Configuration{SupportSources: []SupportSource{}},
		registry:      nil,
	}

	supportSources := []SupportSource{{Identifier: "aboutCloudoguToken", External: false, Href: "/local/href"}, {Identifier: "myCloudogu", External: true, Href: "https://ecosystem.cloudogu.com/"}}

	actual, err := reader.readSupport(supportSources)
	if err != nil {
		t.Fail()
	}

	actualCategories := Categories{{Title: "Support", Entries: []Entry{
		{Title: "aboutCloudoguToken", Target: TARGET_SELF, Href: "/local/href"},
		{Title: "myCloudogu", Target: TARGET_EXTERNAL, Href: "https://ecosystem.cloudogu.com/"}}}}

	assert.Equal(t, actualCategories, actual, "readSupport did not return a correct Category")
}
