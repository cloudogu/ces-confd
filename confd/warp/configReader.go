package warp

import (
	"log"

	"sort"

	"github.com/cloudogu/ces-confd/confd/registry"
	"github.com/pkg/errors"
)

// ConfigReader reads the configuration for the warp menu from etcd
type ConfigReader struct {
	configuration Configuration
	registry      registry.Registry
}

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
	return nil, errors.New("wrong source type")
}

func (reader *ConfigReader) readSupport(supportSources []SupportSource) (Categories, error) {

	var supportEntries []EntryWithCategory

	for _, supportSource := range supportSources {
		// supportSource -> EntryWithCategory
		entry := Entry{}
		if supportSource.external {
			entry = Entry{Title: supportSource.identifier, Target: TARGET_EXTERNAL}
		} else {
			entry = Entry{Title: supportSource.identifier, Target: TARGET_SELF}
		}

		entryWithCategory := EntryWithCategory{Entry: entry, Category: "Support"}
		log.Printf("suppport-entry with Category:%v and Entry %v", entryWithCategory.Category, entryWithCategory.Entry)
		supportEntries = append(supportEntries, entryWithCategory)
	}

	return reader.createCategories(supportEntries), nil
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
	// add support category
	supportCategory, err := reader.readSupport(configuration.SupportSources)
	if err != nil {
		log.Println("Error during support read:", err)
	}
	data.insertCategories(supportCategory)

	return data, nil
}
