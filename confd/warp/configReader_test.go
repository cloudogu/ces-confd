package warp

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v2"
	"log"
	"os"
	"testing"
)

func TestConfigReader_readSupport(t *testing.T) {
	t.Run("should successfully read support entries without filters", func(t *testing.T) {
		supportSources := []SupportSource{
			{Identifier: "aboutCloudoguToken", External: false, Href: "/local/href"},
			{Identifier: "myCloudogu", External: true, Href: "https://ecosystem.cloudogu.com/"},
			{Identifier: "docsCloudoguComUrl", External: true, Href: "https://docs.cloudogu.com/"},
		}

		reader := &ConfigReader{}

		actual := reader.readSupport(supportSources, false, []string{}, []string{})

		expectedCategories := Categories{
			{Title: "Support", Entries: []Entry{
				{Title: "aboutCloudoguToken", Target: TARGET_SELF, Href: "/local/href"},
				{Title: "myCloudogu", Target: TARGET_EXTERNAL, Href: "https://ecosystem.cloudogu.com/"},
				{Title: "docsCloudoguComUrl", Target: TARGET_EXTERNAL, Href: "https://docs.cloudogu.com/"},
			}}}
		assert.Equal(t, expectedCategories, actual)
	})

	t.Run("should block all entries", func(t *testing.T) {
		supportSources := []SupportSource{
			{Identifier: "aboutCloudoguToken", External: false, Href: "/local/href"},
			{Identifier: "myCloudogu", External: true, Href: "https://ecosystem.cloudogu.com/"},
			{Identifier: "docsCloudoguComUrl", External: true, Href: "https://docs.cloudogu.com/"},
		}

		reader := &ConfigReader{}

		actual := reader.readSupport(supportSources, true, []string{}, []string{})

		expectedCategories := Categories{}
		assert.Equal(t, expectedCategories, actual)
	})

	t.Run("should add allowed entries when blocked", func(t *testing.T) {
		supportSources := []SupportSource{
			{Identifier: "aboutCloudoguToken", External: false, Href: "/local/href"},
			{Identifier: "myCloudogu", External: true, Href: "https://ecosystem.cloudogu.com/"},
			{Identifier: "docsCloudoguComUrl", External: true, Href: "https://docs.cloudogu.com/"},
		}

		reader := &ConfigReader{}

		actual := reader.readSupport(supportSources, true, []string{}, []string{"myCloudogu"})

		expectedCategories := Categories{
			{Title: "Support", Entries: []Entry{
				{Title: "myCloudogu", Target: TARGET_EXTERNAL, Href: "https://ecosystem.cloudogu.com/"},
			}}}
		assert.Equal(t, expectedCategories, actual)
	})

	t.Run("should remove disabled entries when not blocked", func(t *testing.T) {
		supportSources := []SupportSource{
			{Identifier: "aboutCloudoguToken", External: false, Href: "/local/href"},
			{Identifier: "myCloudogu", External: true, Href: "https://ecosystem.cloudogu.com/"},
			{Identifier: "docsCloudoguComUrl", External: true, Href: "https://docs.cloudogu.com/"},
		}

		reader := &ConfigReader{}

		actual := reader.readSupport(supportSources, false, []string{"aboutCloudoguToken", "docsCloudoguComUrl"}, []string{})

		expectedCategories := Categories{
			{Title: "Support", Entries: []Entry{
				{Title: "myCloudogu", Target: TARGET_EXTERNAL, Href: "https://ecosystem.cloudogu.com/"},
			}}}
		assert.Equal(t, expectedCategories, actual)
	})

	t.Run("should remove disabled entries when not blocked", func(t *testing.T) {
		supportSources := []SupportSource{
			{Identifier: "aboutCloudoguToken", External: false, Href: "/local/href"},
			{Identifier: "myCloudogu", External: true, Href: "https://ecosystem.cloudogu.com/"},
			{Identifier: "docsCloudoguComUrl", External: true, Href: "https://docs.cloudogu.com/"},
		}

		reader := &ConfigReader{}

		actual := reader.readSupport(supportSources, false, []string{"aboutCloudoguToken", "docsCloudoguComUrl"}, []string{})

		expectedCategories := Categories{
			{Title: "Support", Entries: []Entry{
				{Title: "myCloudogu", Target: TARGET_EXTERNAL, Href: "https://ecosystem.cloudogu.com/"},
			}}}
		assert.Equal(t, expectedCategories, actual)
	})

}

func TestConfigReader_readFromConfig(t *testing.T) {
	t.Run("should read categories from config", func(t *testing.T) {
		mockRegistry := newMockConfigRegistry(t)
		mockRegistry.On("Get", blockWarpSupportCategoryConfigurationKey).
			Return(&client.Response{Node: &client.Node{Value: "false"}}, nil)
		mockRegistry.On("Get", disabledWarpSupportEntriesConfigurationKey).
			Return(&client.Response{Node: &client.Node{Value: "[\"lorem\", \"ipsum\"]"}}, nil)
		mockRegistry.On("Get", allowedWarpSupportEntriesConfigurationKey).
			Return(&client.Response{Node: &client.Node{Value: "[\"lorem\", \"ipsum\"]"}}, nil)
		mockRegistry.On("Get", "/config/externals").
			Return(&client.Response{Node: &client.Node{Nodes: []*client.Node{
				{Key: "/config/externals/ext1"},
			}}}, nil)
		mockRegistry.On("Get", "/config/externals/ext1").
			Return(&client.Response{Node: &client.Node{Value: "{\"DisplayName\": \"ext1\", \"URL\": \"https://my.url/ext1\", \"Description\": \"ext1 Description\", \"Category\": \"Documentation\"}"}}, nil)

		reader := &ConfigReader{
			registry: mockRegistry,
		}

		testSources := []Source{{Path: "/config/externals", SourceType: "externals", Tag: "tag"}}
		testSupportSoureces := []SupportSource{{Identifier: "supportSrc", External: true, Href: "https://support.source"}}

		actual, err := reader.readFromConfig(Configuration{Sources: testSources, SupportSources: testSupportSoureces})
		require.NoError(t, err)

		expectedCategories := Categories{
			{Title: "Documentation", Entries: []Entry{
				{DisplayName: "ext1", Title: "ext1 Description", Target: TARGET_EXTERNAL, Href: "https://my.url/ext1"},
			}},
			{Title: "Support", Entries: []Entry{
				{Title: "supportSrc", Target: TARGET_EXTERNAL, Href: "https://support.source"},
			}},
		}
		assert.Equal(t, expectedCategories, actual)
	})

	t.Run("should log errors when failing to read from registry client", func(t *testing.T) {
		mockRegistry := newMockConfigRegistry(t)
		mockRegistry.On("Get", blockWarpSupportCategoryConfigurationKey).Return(nil, assert.AnError)
		mockRegistry.On("Get", disabledWarpSupportEntriesConfigurationKey).Return(nil, assert.AnError)
		mockRegistry.On("Get", allowedWarpSupportEntriesConfigurationKey).Return(nil, assert.AnError)
		mockRegistry.On("Get", "/config/externals").Return(nil, assert.AnError)

		reader := &ConfigReader{
			registry: mockRegistry,
		}

		testSources := []Source{{Path: "/config/externals", SourceType: "externals", Tag: "tag"}}
		testSupportSoureces := []SupportSource{}

		// capture log
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		actual, err := reader.readFromConfig(Configuration{Sources: testSources, SupportSources: testSupportSoureces})
		require.NoError(t, err)

		assert.Nil(t, actual)

		// assert log
		assert.Contains(t, buf.String(), "Error during read: failed to read root entry /config/externals from etcd")
		assert.Contains(t, buf.String(), "Warning, could not read etcd Key: /config/_global/block_warpmenu_support_category.")
		assert.Contains(t, buf.String(), "Warning, could not read etcd Key: /config/_global/disabled_warpmenu_support_entries.")
		assert.Contains(t, buf.String(), "Warning, could not read etcd Key: /config/_global/allowed_warpmenu_support_entries.")
	})

}

func TestConfigReader_readStrings(t *testing.T) {
	t.Run("should successfully read strings", func(t *testing.T) {
		response := client.Response{Node: &client.Node{Value: "[\"lorem\", \"ipsum\"]"}}
		mockRegistry := newMockConfigRegistry(t)
		mockRegistry.On("Get", "/config/_global/disabled_warpmenu_support_entries").Return(&response, nil)

		reader := &ConfigReader{
			registry: mockRegistry,
		}

		identifiers, err := reader.readStrings("/config/_global/disabled_warpmenu_support_entries")
		require.NoError(t, err)
		assert.Equal(t, []string{"lorem", "ipsum"}, identifiers)
	})

	t.Run("should fail reading from registry", func(t *testing.T) {
		mockRegistry := newMockConfigRegistry(t)
		mockRegistry.On("Get", "/config/_global/disabled_warpmenu_support_entries").Return(nil, assert.AnError)

		reader := &ConfigReader{
			registry: mockRegistry,
		}

		identifiers, err := reader.readStrings("/config/_global/disabled_warpmenu_support_entries")
		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to read configuration entry /config/_global/disabled_warpmenu_support_entries from etcd")
		assert.Equal(t, []string{}, identifiers)
	})

	t.Run("should fail unmarshalling", func(t *testing.T) {
		response := client.Response{Node: &client.Node{Value: "not-a-string-array 123"}}
		mockRegistry := newMockConfigRegistry(t)
		mockRegistry.On("Get", "/config/_global/disabled_warpmenu_support_entries").Return(&response, nil)

		reader := &ConfigReader{
			registry: mockRegistry,
		}

		identifiers, err := reader.readStrings("/config/_global/disabled_warpmenu_support_entries")
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to unmarshal etcd key to string slice:")
		assert.Equal(t, []string{}, identifiers)
	})
}

func TestConfigReader_readBool(t *testing.T) {
	t.Run("should successfully read true bool", func(t *testing.T) {
		response := client.Response{Node: &client.Node{Value: "true"}}
		mockRegistry := newMockConfigRegistry(t)
		mockRegistry.On("Get", "/config/_global/myBool").Return(&response, nil)

		reader := &ConfigReader{
			registry: mockRegistry,
		}

		boolValue, err := reader.readBool("/config/_global/myBool")
		require.NoError(t, err)
		assert.True(t, boolValue)
	})

	t.Run("should successfully read false bool", func(t *testing.T) {
		response := client.Response{Node: &client.Node{Value: "false"}}
		mockRegistry := newMockConfigRegistry(t)
		mockRegistry.On("Get", "/config/_global/myBool").Return(&response, nil)

		reader := &ConfigReader{
			registry: mockRegistry,
		}

		boolValue, err := reader.readBool("/config/_global/myBool")
		require.NoError(t, err)
		assert.False(t, boolValue)
	})

	t.Run("should fail reading from registry", func(t *testing.T) {
		mockRegistry := newMockConfigRegistry(t)
		mockRegistry.On("Get", "/config/_global/myBool").Return(nil, assert.AnError)

		reader := &ConfigReader{
			registry: mockRegistry,
		}

		boolValue, err := reader.readBool("/config/_global/myBool")
		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to read configuration entry /config/_global/myBool from etcd")
		assert.False(t, boolValue)
	})

	t.Run("should fail unmarshalling", func(t *testing.T) {
		response := client.Response{Node: &client.Node{Value: "not a bool"}}
		mockRegistry := newMockConfigRegistry(t)
		mockRegistry.On("Get", "/config/_global/myBool").Return(&response, nil)

		reader := &ConfigReader{
			registry: mockRegistry,
		}

		boolValue, err := reader.readBool("/config/_global/myBool")
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to unmarshal etcd key to bool:")
		assert.False(t, boolValue)
	})
}
