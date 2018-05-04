package warp

import (
	"encoding/json"

	"github.com/cloudogu/ces-confd/confd/registry"
	"github.com/pkg/errors"
)

type externalEntry struct {
	DisplayName string
	URL         string
	Description string
	Category    string
}

func readAndUnmarshalExternal(registry registry.Registry, key string) (EntryWithCategory, error) {
	externalBytes, err := readExternalAsBytes(registry, key)
	if err != nil {
		return EntryWithCategory{}, nil
	}

	return unmarshalExternal(externalBytes)
}

func readExternalAsBytes(registry registry.Registry, key string) ([]byte, error) {
	resp, err := registry.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read key %s from etcd", key)
	}

	return []byte(resp.Node.Value), nil
}

func unmarshalExternal(externalBytes []byte) (EntryWithCategory, error) {
	externalEntry := externalEntry{}
	err := json.Unmarshal([]byte(externalBytes), &externalEntry)
	if err != nil {
		return EntryWithCategory{}, errors.Wrap(err, "failed to unmarshall external")
	}

	return mapExternalEntry(externalEntry)
}

func mapExternalEntry(entry externalEntry) (EntryWithCategory, error) {
	if entry.DisplayName == "" {
		return EntryWithCategory{}, errors.New("could not find DisplayName on external entry")
	}
	if entry.URL == "" {
		return EntryWithCategory{}, errors.New("could not find URL on external entry")
	}
	if entry.Category == "" {
		return EntryWithCategory{}, errors.New("could not find Category on external entry")
	}
	return EntryWithCategory{
		Entry: Entry{
			DisplayName: entry.DisplayName,
			Title:       entry.Description,
			Href:        entry.URL,
			Target:      TARGET_EXTERNAL,
		},
		Category: entry.Category,
	}, nil
}
