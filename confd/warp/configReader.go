package warp

import (
	"context"
	"log"

	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
)

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
