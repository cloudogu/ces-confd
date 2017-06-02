package warp

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"log"

	. "github.com/cloudogu/ces-confd/confd"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"

	"io/ioutil"
)

type Configuration struct {
	Sources []Source
	Target  string
	Order   Order
}

type Source struct {
	Path       string
	SourceType string `yaml:"type"`
	Tag        string
}

// Category category of multiple entries in the warp menu
type Category struct {
	Title   string
	Order   int
	Entries Entries
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
	Target      string
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

func unmarshalDogu(kapi client.KeysAPI, key string) (RawData, error) {
	resp, err := kapi.Get(context.Background(), key+"/current", nil)
	if err != nil {
		// the dogu seems to be unregistered
		if isKeyNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to read key %s from etcd", key)
	}

	version := resp.Node.Value
	resp, err = kapi.Get(context.Background(), key+"/"+version, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read version child from key %s", key)
	}

	dogu := RawData{}
	err = json.Unmarshal([]byte(resp.Node.Value), &dogu)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall json from etcd")
	}

	return dogu, nil
}

func unmarshalExternal(kapi client.KeysAPI, key string) (RawData, error) {
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read key %s from etcd", key)
	}
	external := RawData{}
	err = json.Unmarshal([]byte(resp.Node.Value), &external)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall json from etcd")
	}

	return external, nil
}

func createHref(dogu RawData) string {
	// remove namespace
	parts := strings.Split(dogu["Name"].(string), "/")
	return "/" + parts[len(parts)-1]
}

func convertElementsToCategories(order Order, elements []RawData, createEntry func(RawData) Entry) Categories {
	categories := map[string]*Category{}

	for _, element := range elements {
		categoryName := element["Category"].(string)
		category := categories[categoryName]
		if category == nil {
			category = &Category{
				Title:   categoryName,
				Entries: Entries{},
				// TODO read order boost from etcd
				Order: order[categoryName],
			}
			categories[categoryName] = category
		}
		warpEntry := createEntry(element)
		category.Entries = append(category.Entries, warpEntry)
	}

	result := Categories{}
	for _, cat := range categories {
		sort.Sort(cat.Entries)
		result = append(result, cat)
	}
	sort.Sort(result)
	return result
}

func createDoguEntry(element RawData) Entry {
	return Entry{
		DisplayName: element["DisplayName"].(string),
		Href:        createHref(element),
		Title:       element["Description"].(string),
		Target:      "self",
	}
}

func createExternalEntry(element RawData) Entry {
	return Entry{
		DisplayName: element["DisplayName"].(string),
		Href:        element["URL"].(string),
		Title:       element["Description"].(string),
		Target:      "external",
	}
}

func filterByTag(dogus []RawData, tag string) []RawData {
	filtered := []RawData{}
	for _, raw := range dogus {
		if raw["Tags"] != nil {
			tags := raw["Tags"].([]interface{})
			if tags != nil && Contains(tags, tag) {
				filtered = append(filtered, raw)
			}
		}
	}
	return filtered
}

// DogusReader reads from etcd and converts the keys and values to a warp menu
// conform structure
func dogusReader(source Source, kapi client.KeysAPI, order Order) (Categories, error) {
	resp, err := kapi.Get(context.Background(), source.Path, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read root entry %s from etcd", source.Path)
	}
	dogus := []RawData{}
	for _, child := range resp.Node.Nodes {
		dogu, err := unmarshalDogu(kapi, child.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshall node from etcd")
		} else if dogu != nil {
			dogus = append(dogus, dogu)
		}
	}

	if source.Tag != "" {
		dogus = filterByTag(dogus, source.Tag)
	}
	return convertElementsToCategories(order, dogus, createDoguEntry), nil
}

func externalsReader(source Source, kapi client.KeysAPI, order Order) (Categories, error) {
	resp, err := kapi.Get(context.Background(), source.Path, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read root entry %s from etcd", source.Path)
	}
	externals := []RawData{}
	for _, child := range resp.Node.Nodes {
		external, err := unmarshalExternal(kapi, child.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshall node from etcd")
		} else if external != nil {
			externals = append(externals, external)
		}
	}
	return convertElementsToCategories(order, externals, createExternalEntry), nil
}

func readSource(source Source, kapi client.KeysAPI, order Order) (Categories, error) {
	switch source.SourceType {
	case "dogus":
		return dogusReader(source, kapi, order)
	case "externals":
		return externalsReader(source, kapi, order)
	}
	return nil, errors.New("wrong source type")
}

func readFromConfig(configuration Configuration, kapi client.KeysAPI) (Categories, error) {
	var data Categories
	for _, source := range configuration.Sources {
		categories, err := readSource(source, kapi, configuration.Order)
		if err != nil {
			log.Println("Error durring read", err)
		}
		data.insertCategories(categories)
	}
	return data, nil
}

func (destination *Categories) insertCategories(newCategories Categories) {
	for _, newCategory := range newCategories {
		destination.insertCategory(newCategory)
	}
}

func (destination *Categories) insertCategory(newCategory *Category) {
	for _, category := range *destination {
		if category.Title == newCategory.Title {
			category.Entries = append(category.Entries, newCategory.Entries...)
			return
		}
	}
	*destination = append(*destination, newCategory)
}

// JSONWriter converts the data to a json
func JSONWriter(target string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal data to json")
	}

	return ioutil.WriteFile(target, bytes, 0755)
}

func execute(configuration Configuration, kapi client.KeysAPI) {
	categories, err := readFromConfig(configuration, kapi)
	if err != nil {
		log.Println("Error durring read", err)
	}
	log.Printf("all found categories: %i", categories)
	JSONWriter(configuration.Target, categories)
}

func watch(source Source, kapi client.KeysAPI, execChannel chan Source) {
	watcherOpts := client.WatcherOptions{AfterIndex: 0, Recursive: true}
	watcher := kapi.Watcher(source.Path, &watcherOpts)
	for {
		resp, err := watcher.Next(context.Background())
		if err != nil {
			// TODO: execute before watch start again? wait to reduce load, in case of unrecoverable error?
			watch(source, kapi, execChannel)
		} else {
			action := resp.Action
			log.Printf("%s changed, action=%s", resp.Node.Key, action)
			execChannel <- source
		}
	}
}

func Run(configuration Configuration, kapi client.KeysAPI, syncWaitGroup sync.WaitGroup) {
	execute(configuration, kapi)
	log.Println("start watcher for warp entries")
	execChannel := make(chan Source)
	for _, source := range configuration.Sources {
		go watch(source, kapi, execChannel)
	}

	for range execChannel {
		execute(configuration, kapi)
	}

	syncWaitGroup.Done()
}
