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

func (r *EtcdRegistry) updateIndexIfNecessary(index uint64) {
	if r.recentIndex == 0 {
		r.indexMutex.Lock()
		if r.recentIndex == 0 {
			r.recentIndex = index
		}
		r.indexMutex.Unlock()
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
			r.recentIndex = 0
			r.indexMutex.Unlock()
			return
		}

		eventChannel <- resp
	}

}
