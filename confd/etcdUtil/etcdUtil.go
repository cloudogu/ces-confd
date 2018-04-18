package etcdUtil

import (
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// GetLastIndex returns the last set index for the provided key
func GetLastIndex(key string, kapi client.KeysAPI) (uint64, error) {
	resp, err := kapi.Get(context.Background(), key, nil)

	if err != nil {
		return 1, errors.Wrap(err, "Could not get index")
	}

	return resp.Index, nil
}
