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

	supportSources := []SupportSource{{Identifier: "aboutCloudoguToken", External: false, Href: "/local/href"}, {Identifier: "myCloudogu", External: true, Href: "https://ecosystem.cloudogu.com/"}, {Identifier: "docsCloudoguComUrl", External: true, Href: "https://docs.cloudogu.com/"}}

	actual, err := reader.readSupport(supportSources, []string{"docsCloudoguComUrl"})

	if err != nil {
		t.Fail()
	}

	expectedCategories := Categories{{Title: "Support", Entries: []Entry{
		{Title: "aboutCloudoguToken", Target: TARGET_SELF, Href: "/local/href"},
		{Title: "myCloudogu", Target: TARGET_EXTERNAL, Href: "https://ecosystem.cloudogu.com/"}}}}

	assert.Equal(t, expectedCategories, actual, "readSupport did not return a correct Category")
}
