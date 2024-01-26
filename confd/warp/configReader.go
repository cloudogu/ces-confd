package warp

import (
	"encoding/json"
	"fmt"
	"github.com/cloudogu/ces-confd/confd/util"
	"log"
	"strconv"

	"sort"

	"github.com/cloudogu/ces-confd/confd/registry"
	"github.com/pkg/errors"
)

// ConfigReader reads the configuration for the warp menu from etcd
type ConfigReader struct {
	configuration Configuration
	registry      registry.Registry
}

type DisabledSupportEntries struct {
	name []string
}

const blockWarpSupportCategoryConfigurationKey = "/config/_global/block_warpmenu_support_category"
const disabledWarpSupportEntriesConfigurationKey = "/config/_global/disabled_warpmenu_support_entries"
const allowedWarpSupportEntriesConfigurationKey = "/config/_global/allowed_warpmenu_support_entries"

func (reader *ConfigReader) createCategories(entries []EntryWithCategory) Categories {
	categories := map[string]*Category{}

	for _, entry := range entries {
		categoryName := entry.Category
		category := categories[categoryName]
		if category == nil {
			category = &Category{
				Title:   categoryName,
				Entries: Entries{},
				// TODO read order boost from etcd
				Order: reader.configuration.Order[categoryName],
			}
			categories[categoryName] = category
		}
		category.Entries = append(category.Entries, entry.Entry)
	}

	result := Categories{}
	for _, cat := range categories {
		sort.Sort(cat.Entries)
		result = append(result, cat)
	}
	sort.Sort(result)
	return result
}

// dogusReader reads from etcd and converts the keys and values to a warp menu
// conform structure
func (reader *ConfigReader) dogusReader(source Source) (Categories, error) {
	log.Printf("read dogus from %s for warp menu", source.Path)
	resp, err := reader.registry.Get(source.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read root entry %s from etcd", source.Path)
	}
	dogus := []EntryWithCategory{}
	for _, child := range resp.Node.Nodes {
		dogu, err := readAndUnmarshalDogu(reader.registry, child.Key, source.Tag)
		if err != nil {
			log.Printf("failed to read and unmarshal dogu: %v", err)
		} else if dogu.Entry.Title != "" { // TODO more explicit way to handle filtered entries
			dogus = append(dogus, dogu)
		}
	}

	return reader.createCategories(dogus), nil
}

func (reader *ConfigReader) externalsReader(source Source) (Categories, error) {
	log.Printf("read externals from %s for warp menu", source.Path)
	resp, err := reader.registry.Get(source.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read root entry %s from etcd", source.Path)
	}
	externals := []EntryWithCategory{}
	for _, child := range resp.Node.Nodes {
		external, err := readAndUnmarshalExternal(reader.registry, child.Key)
		if err != nil {
			log.Printf("failed to read and unmarshal external: %v", err)
		} else {
			externals = append(externals, external)
		}
	}
	return reader.createCategories(externals), nil
}

func (reader *ConfigReader) readSource(source Source) (Categories, error) {
	switch source.SourceType {
	case "dogus":
		return reader.dogusReader(source)
	case "externals":
		return reader.externalsReader(source)
	}
	return nil, errors.Errorf("wrong source type: %v", source.SourceType)
}

func (reader *ConfigReader) readStrings(registryKey string) ([]string, error) {
	entry, err := reader.registry.Get(registryKey)
	if err != nil {
		return []string{}, fmt.Errorf("failed to read configuration entry %s from etcd: %w", registryKey, err)
	}

	var strings []string
	err = json.Unmarshal([]byte(entry.Node.Value), &strings)
	if err != nil {
		return []string{}, fmt.Errorf("failed to unmarshal etcd key to string slice: %w", err)
	}

	return strings, nil
}

func (reader *ConfigReader) readBool(registryKey string) (bool, error) {
	entry, err := reader.registry.Get(registryKey)
	if err != nil {
		return false, fmt.Errorf("failed to read configuration entry %s from etcd: %w", registryKey, err)
	}

	boolValue, err := strconv.ParseBool(entry.Node.Value)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal etcd key to bool: %w", err)
	}

	return boolValue, nil
}

func (reader *ConfigReader) readSupport(supportSources []SupportSource, blocked bool, disabledEntries []string, allowedEntries []string) Categories {
	var supportEntries []EntryWithCategory

	for _, supportSource := range supportSources {
		if (blocked && util.StringInSlice(supportSource.Identifier, allowedEntries)) || (!blocked && !util.StringInSlice(supportSource.Identifier, disabledEntries)) {
			// support category is blocked, but this entry is explicitly allowed OR support category is NOT blocked and this entry is NOT explicitly disabled

			entry := Entry{Title: supportSource.Identifier, Href: supportSource.Href, Target: TARGET_SELF}
			if supportSource.External {
				entry.Target = TARGET_EXTERNAL
			}

			supportEntries = append(supportEntries, EntryWithCategory{Entry: entry, Category: "Support"})
		}
	}

	return reader.createCategories(supportEntries)
}

func (reader *ConfigReader) readFromConfig(configuration Configuration) (Categories, error) {
	var data Categories

	for _, source := range configuration.Sources {
		categories, err := reader.readSource(source)
		if err != nil {
			log.Println("Error during read:", err)
		}
		data.insertCategories(categories)
	}

	log.Println("read SupportEntries")

	isSupportCategoryBlocked, err := reader.readBool(blockWarpSupportCategoryConfigurationKey)
	if err != nil {
		log.Printf("Warning, could not read etcd Key: %v. Err: %v", blockWarpSupportCategoryConfigurationKey, err)
	}

	disabledSupportEntries, err := reader.readStrings(disabledWarpSupportEntriesConfigurationKey)
	if err != nil {
		log.Printf("Warning, could not read etcd Key: %v. Err: %v", disabledWarpSupportEntriesConfigurationKey, err)
	}

	allowedSupportEntries, err := reader.readStrings(allowedWarpSupportEntriesConfigurationKey)
	if err != nil {
		log.Printf("Warning, could not read etcd Key: %v. Err: %v", allowedWarpSupportEntriesConfigurationKey, err)
	}

	// add support category
	supportCategory := reader.readSupport(configuration.SupportSources, isSupportCategoryBlocked, disabledSupportEntries, allowedSupportEntries)

	if supportCategory.Len() == 0 {
		log.Printf("support category is empty, no support category will be added to menu.json")
		return data, nil
	}

	data.insertCategories(supportCategory)
	return data, nil
}
