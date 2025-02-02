package warp

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/cloudogu/ces-confd/confd"
	"github.com/cloudogu/ces-confd/confd/registry"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/client/v2"
)

// Configuration for warp menu creation
type Configuration struct {
	Sources        []Source
	Target         string
	Order          confd.Order
	SupportSources []SupportSource `yaml:"support"`
}

// Source in etcd
type Source struct {
	Path       string
	SourceType string `yaml:"type"`
	Tag        string
}

// Source for SupportEntries from yaml
type SupportSource struct {
	Identifier string `yaml:"identifier"`
	External   bool   `yaml:"external"`
	Href       string `yaml:"href"`
}

// Category category of multiple entries in the warp menu
type Category struct {
	Title   string
	Order   int
	Entries Entries
}

// String returns the title of the category
func (category Category) String() string {
	return category.Title
}

// Categories collection of warp categories
type Categories []*Category

// sort methods

func (categories Categories) Len() int {
	return len(categories)
}

func (categories Categories) Less(i, j int) bool {
	if categories[i].Order == categories[j].Order {
		return categories[i].Title < categories[j].Title
	}
	return categories[i].Order > categories[j].Order
}

func (categories Categories) Swap(i, j int) {
	categories[i], categories[j] = categories[j], categories[i]
}

// Entry link in the warp menu
type Entry struct {
	DisplayName string
	Href        string
	Title       string
	Target      Target
}

// Target defines the target of the link
type Target uint8

const (
	// TARGET_SELF means the link is part of the internal system
	TARGET_SELF Target = iota + 1
	// TARGET_EXTERNAL link is outside from the system
	TARGET_EXTERNAL
)

func (target Target) MarshalJSON() ([]byte, error) {
	switch target {
	case TARGET_SELF:
		return target.asJSONString("self"), nil
	case TARGET_EXTERNAL:
		return target.asJSONString("external"), nil
	default:
		return nil, errors.Errorf("unknow target type %d", target)
	}
}

func (target Target) asJSONString(value string) []byte {
	return []byte("\"" + value + "\"")
}

// EntryWithCategory is a dto for entries with a category
type EntryWithCategory struct {
	Entry    Entry
	Category string
}

// Entries is a collection of warp entries
type Entries []Entry

// sort methods

func (entries Entries) Len() int {
	return len(entries)
}

func (entries Entries) Less(i, j int) bool {
	return entries[i].DisplayName < entries[j].DisplayName
}

func (entries Entries) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}

func isKeyNotFound(err error) bool {
	if cErr, ok := err.(client.Error); ok {
		return cErr.Code == client.ErrorCodeKeyNotFound
	}
	return false
}

func (categories *Categories) insertCategories(newCategories Categories) {
	for _, newCategory := range newCategories {
		categories.insertCategory(newCategory)
	}
}

func (categories *Categories) insertCategory(newCategory *Category) {
	for _, category := range *categories {
		if category.Title == newCategory.Title {
			category.Entries = append(category.Entries, newCategory.Entries...)
			return
		}
	}
	*categories = append(*categories, newCategory)
}

// JSONWriter converts the data to a json
func jsonWriter(target string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal data to json")
	}

	return ioutil.WriteFile(target, bytes, 0755)
}

func execute(configuration Configuration, registry registry.Registry) {
	reader := ConfigReader{
		registry:      registry,
		configuration: configuration,
	}
	categories, err := reader.readFromConfig(configuration)
	if err != nil {
		log.Println("Error during read:", err)
		return
	}
	log.Printf("all found categories: %v", categories)
	err = jsonWriter(configuration.Target, categories)
	if err != nil {
		log.Printf("failed to write warp menu as json: %v", err)
	}
}

// Run creates the warp menu and update the menu whenever a relevant etcd key was changed
func Run(configuration Configuration, registry registry.Registry) {

	log.Println("start watcher for warp entries")
	warpChannel := make(chan *client.Response)

	for _, source := range configuration.Sources {

		go func(source Source) {
			for {
				execute(configuration, registry)
				registry.Watch(source.Path, true, warpChannel)
			}
		}(source)
	}

	for range warpChannel {
		execute(configuration, registry)
	}

}
