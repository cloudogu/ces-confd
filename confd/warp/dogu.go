package warp

import (
	"encoding/json"

	"strings"

	"github.com/cloudogu/ces-confd/confd"
	"github.com/pkg/errors"
)

type doguEntry struct {
	Name        string
	DisplayName string
	Description string
	Category    string
	Tags        []string
}

func readAndUnmarshalDogu(registry configRegistry, key string, tag string) (EntryWithCategory, error) {
	doguBytes, err := readDoguAsBytes(registry, key)
	if err != nil {
		return EntryWithCategory{}, err
	}

	doguEntry, err := unmarshalDogu(doguBytes)
	if err != nil {
		return EntryWithCategory{}, err
	}

	if tag == "" || confd.ContainsString(doguEntry.Tags, tag) {
		return mapDoguEntry(doguEntry)
	}

	// TODO more explicit way to handle filtered entries
	return EntryWithCategory{}, nil
}

func readDoguAsBytes(registry configRegistry, key string) ([]byte, error) {
	resp, err := registry.Get(key + "/current")
	if err != nil {
		// the dogu seems to be unregistered
		if isKeyNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to read key %s from etcd", key)
	}

	version := resp.Node.Value
	resp, err = registry.Get(key + "/" + version)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read version child from key %s", key)
	}

	return []byte(resp.Node.Value), nil
}

func unmarshalDogu(doguBytes []byte) (doguEntry, error) {
	doguEntry := doguEntry{}
	err := json.Unmarshal([]byte(doguBytes), &doguEntry)
	if err != nil {
		return doguEntry, errors.Wrap(err, "failed to unmarshall json from etcd")
	}
	return doguEntry, nil
}

func mapDoguEntry(entry doguEntry) (EntryWithCategory, error) {
	if entry.Name == "" {
		return EntryWithCategory{}, errors.New("name is required for dogu entries")
	}

	displayName := entry.DisplayName
	if displayName == "" {
		displayName = entry.Name
	}

	return EntryWithCategory{
		Entry: Entry{
			DisplayName: displayName,
			Title:       entry.Description,
			Target:      TARGET_SELF,
			Href:        createDoguHref(entry.Name),
		},
		Category: entry.Category,
	}, nil
}

func createDoguHref(name string) string {
	// remove namespace
	parts := strings.Split(name, "/")
	return "/" + parts[len(parts)-1]
}
