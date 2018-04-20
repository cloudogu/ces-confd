package registry

import (
	"log"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// EtcdRegistry implements the Registry interface for etcd
type EtcdRegistry struct {
	keysAPI     client.KeysAPI
	recentIndex uint64
}

// NewRegistry creates and configures a new EtcdRegistry
func NewRegistry(config Config) (*EtcdRegistry, error) {
	cfg := client.Config{
		Endpoints: config.Endpoints,
	}
	c, err := client.New(cfg)

	if err != nil {
		log.Printf("Could not create client: %v", err)
		return nil, err
	}
	keysAPI := client.NewKeysAPI(c)
	return &EtcdRegistry{keysAPI, 0}, nil
}

// Get returns the value associated with the provided key
func (r EtcdRegistry) Get(key string) (*client.Response, error) {
	resp, err := r.keysAPI.Get(context.Background(), key, nil)

	if err != nil {
		log.Printf("Error getting value for key %s: %v", key, err)
		return resp, err
	}
	r.recentIndex = resp.Index
	return resp, nil
}

// Watch watches for changes of the provided key and sends the event through the channel
func (r EtcdRegistry) Watch(key string, recursive bool, eventChannel chan *client.Response) {
	options := client.WatcherOptions{AfterIndex: r.recentIndex, Recursive: recursive}
	watcher := r.keysAPI.Watcher(key, &options)

	go func() {
		for {
			resp, err := watcher.Next(context.Background())

			if err != nil {
				log.Printf("Could not get event: %v", err)
				r.Watch(key, recursive, eventChannel)
				return
			}

			//			event := Event{Action: resp.Action, Value: resp.Node.Value}
			eventChannel <- resp
			r.recentIndex = resp.Index
		}
	}()

}
