package registry

import (
	"log"
	"sync"

	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// EtcdRegistry implements the Registry interface for etcd
type EtcdRegistry struct {
	keysAPI     client.KeysAPI
	indexMutex  sync.Mutex
	recentIndex uint64
}

// NewEtcdRegistry creates and configures a new EtcdRegistry
func NewEtcdRegistry(config Config) (*EtcdRegistry, error) {
	cfg := client.Config{
		Endpoints: config.Endpoints,
	}
	c, err := client.New(cfg)

	if err != nil {
		errors.Wrapf(err, "Could not create client: %v")
		return nil, err
	}
	keysAPI := client.NewKeysAPI(c)
	return &EtcdRegistry{keysAPI: keysAPI, recentIndex: 0}, nil
}

// Get returns the value associated with the provided key
func (r *EtcdRegistry) Get(key string) (*client.Response, error) {
	resp, err := r.keysAPI.Get(context.Background(), key, nil)

	if err != nil {
		errors.Wrapf(err, "Error getting value for key %s:", key)
		return resp, err
	}

	r.updateIndexIfNecessary(resp.Index)
	return resp, nil
}

// We only update the recent index iff it is 0; which happens only in 2 cases:
// 1. At startup
// 2. In case of an error during watch
// We do this, to not miss any changes made to etcd between
// 1. Startup and starting the watcher
// 2. An error and the restart of the watcher
func (r *EtcdRegistry) updateIndexIfNecessary(index uint64) {
	if r.recentIndex == 0 {
		r.indexMutex.Lock()
		defer r.indexMutex.Unlock()
		if r.recentIndex == 0 {
			r.recentIndex = index
		}

	}
}

// Watch watches for changes of the provided key and sends the event through the channel
func (r *EtcdRegistry) Watch(key string, recursive bool, eventChannel chan *client.Response) {

	options := client.WatcherOptions{AfterIndex: r.recentIndex, Recursive: recursive}
	watcher := r.keysAPI.Watcher(key, &options)
	for {
		resp, err := watcher.Next(context.Background())

		if err != nil {
			log.Printf("Could not get event: %v", err)
			r.indexMutex.Lock()
			defer r.indexMutex.Unlock()
			r.recentIndex = 0
			return
		}

		eventChannel <- resp
	}

}
