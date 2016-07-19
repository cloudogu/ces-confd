package main

import (
	"encoding/json"
	"sort"
	"strings"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
)

// WarpCategory category of multiple entries in the warp menu
type WarpCategory struct {
	Title   string
	Order   int
	Entries WarpEntries
}

// WarpCategories collection of warp categories
type WarpCategories []*WarpCategory

// sort methods

func (categories WarpCategories) Len() int {
	return len(categories)
}

func (categories WarpCategories) Less(i, j int) bool {
	if categories[i].Order == categories[j].Order {
		return categories[i].Title < categories[j].Title
	}
	return categories[i].Order > categories[j].Order
}

func (categories WarpCategories) Swap(i, j int) {
	categories[i], categories[j] = categories[j], categories[i]
}

// WarpEntry link in the warp menu
type WarpEntry struct {
	DisplayName string
	Href        string
	Title       string
}

// WarpEntries is a collection of warp entries
type WarpEntries []WarpEntry

// sort methods

func (entries WarpEntries) Len() int {
	return len(entries)
}

func (entries WarpEntries) Less(i, j int) bool {
	return entries[i].DisplayName < entries[j].DisplayName
}

func (entries WarpEntries) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}

func isKeyNotFound(err error) bool {
	if cErr, ok := err.(client.Error); ok {
		return cErr.Code == client.ErrorCodeKeyNotFound
	}
	return false
}

func unmarshal(kapi client.KeysAPI, key string) (RawData, error) {
	resp, err := kapi.Get(context.Background(), key+"/current", nil)
	if err != nil {
		// the dogu seems to be unregistered
		if isKeyNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	version := resp.Node.Value
	resp, err = kapi.Get(context.Background(), key+"/"+version, nil)
	if err != nil {
		return nil, err
	}

	dogu := RawData{}
	err = json.Unmarshal([]byte(resp.Node.Value), &dogu)
	if err != nil {
		return nil, err
	}

	return dogu, nil
}

func createHref(dogu RawData) string {
	// remove namespace
	parts := strings.Split(dogu["Name"].(string), "/")
	return "/" + parts[len(parts)-1]
}

func convert(entry Entry, dogus []RawData) WarpCategories {
	categories := map[string]*WarpCategory{}

	for _, dogu := range dogus {
		categoryName := dogu["Category"].(string)
		category := categories[categoryName]
		if category == nil {
			category = &WarpCategory{
				Title:   categoryName,
				Entries: WarpEntries{},
				// TODO read order boost from etcd
				Order: entry.Order[categoryName],
			}
			categories[categoryName] = category
		}

		category.Entries = append(category.Entries, WarpEntry{
			DisplayName: dogu["DisplayName"].(string),
			Href:        createHref(dogu),
			Title:       dogu["Description"].(string),
		})
	}

	result := WarpCategories{}
	for _, cat := range categories {
		sort.Sort(cat.Entries)
		result = append(result, cat)
	}
	sort.Sort(result)
	return result
}

func filterByTag(dogus []RawData, tag string) []RawData {
	filtered := []RawData{}
	for _, raw := range dogus {
		if raw["Tags"] != nil {
			tags := raw["Tags"].([]interface{})
			if tags != nil && contains(tags, tag) {
				filtered = append(filtered, raw)
			}
		}
	}
	return filtered
}

// WarpReader reads from etcd and converts the keys and values to a warp menu
// conform structure
func WarpReader(kapi client.KeysAPI, entry Entry, root string) (interface{}, error) {
	resp, err := kapi.Get(context.Background(), root, nil)
	if err != nil {
		return nil, err
	}

	dogus := []RawData{}
	for _, child := range resp.Node.Nodes {
		dogu, err := unmarshal(kapi, child.Key)
		if err != nil {
			return nil, err
		} else if dogu != nil {
			dogus = append(dogus, dogu)
		}
	}

	if entry.Tag != "" {
		dogus = filterByTag(dogus, entry.Tag)
	}

	return convert(entry, dogus), nil
}
