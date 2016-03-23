package main

import (
	"encoding/json"
	"sort"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)

// WarpCategory category of multiple entries in the warp menu
type WarpCategory struct {
	Title   string
	Entries WarpEntries
}

// WarpCategories collection of warp categories
type WarpCategories []*WarpCategory

// sort methods

func (categories WarpCategories) Len() int {
	return len(categories)
}

func (categories WarpCategories) Less(i, j int) bool {
	return categories[i].Title < categories[j].Title
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

func unmarshal(kapi client.KeysAPI, key string) (RawData, error) {
	resp, err := kapi.Get(context.Background(), key+"/current", nil)
	if err != nil {
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

func convert(dogus []RawData) WarpCategories {
	categories := map[string]*WarpCategory{}

	for _, dogu := range dogus {
		categoryName := dogu["Category"].(string)
		category := categories[categoryName]
		if category == nil {
			category = &WarpCategory{
				Title:   categoryName,
				Entries: WarpEntries{},
			}
			categories[categoryName] = category
		}

		category.Entries = append(category.Entries, WarpEntry{
			DisplayName: dogu["DisplayName"].(string),
			Href:        "/" + dogu["Name"].(string),
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

// WarpReader reads from etcd and converts the keys and values to a warp menu
// conform structure
func WarpReader(kapi client.KeysAPI, root string) (interface{}, error) {
	resp, err := kapi.Get(context.Background(), root, nil)
	if err != nil {
		return nil, err
	}

	dogus := []RawData{}
	for _, child := range resp.Node.Nodes {
		dogu, err := unmarshal(kapi, child.Key)
		if err != nil {
			return nil, err
		}
		dogus = append(dogus, dogu)
	}

	return convert(dogus), nil
}
