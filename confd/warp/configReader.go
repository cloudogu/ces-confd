package warp

import (
	"context"
	"log"

	"sort"

	. "github.com/cloudogu/ces-confd/confd"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
)

type ConfigReader struct {
	configuration Configuration
	kapi          client.KeysAPI
}

func (reader *ConfigReader) convertElementsToCategories(elements []RawData, createEntry func(RawData) Entry) Categories {
	categories := map[string]*Category{}

	for _, element := range elements {
		categoryName := element["Category"].(string)
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

// dogusReader reads from etcd and converts the keys and values to a warp menu
// conform structure
func (reader *ConfigReader) dogusReader(source Source) (Categories, error) {
	resp, err := reader.kapi.Get(context.Background(), source.Path, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read root entry %s from etcd", source.Path)
	}
	dogus := []RawData{}
	for _, child := range resp.Node.Nodes {
		dogu, err := unmarshalDogu(reader.kapi, child.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshall node from etcd")
		} else if dogu != nil {
			dogus = append(dogus, dogu)
		}
	}

	if source.Tag != "" {
		dogus = filterByTag(dogus, source.Tag)
	}
	return reader.convertElementsToCategories(dogus, createDoguEntry), nil
}

func (reader *ConfigReader) externalsReader(source Source) (Categories, error) {
	resp, err := reader.kapi.Get(context.Background(), source.Path, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read root entry %s from etcd", source.Path)
	}
	externals := []RawData{}
	for _, child := range resp.Node.Nodes {
		external, err := unmarshalExternal(reader.kapi, child.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshall node from etcd")
		} else if external != nil {
			externals = append(externals, external)
		}
	}
	return reader.convertElementsToCategories(externals, createExternalEntry), nil
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

func (reader *ConfigReader) readFromConfig(configuration Configuration, kapi client.KeysAPI) (Categories, error) {
	var data Categories
	for _, source := range configuration.Sources {
		categories, err := reader.readSource(source)
		if err != nil {
			log.Println("Error durring read", err)
		}
		data.insertCategories(categories)
	}
	return data, nil
}
