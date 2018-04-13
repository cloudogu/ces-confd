package etcdUtil

import (
  "github.com/coreos/etcd/client"
  "log"
  "golang.org/x/net/context"
)

func GetLastIndex(key string, kapi client.KeysAPI) (uint64, error) {
  resp, err := kapi.Get(context.Background(), key, nil)

  if err != nil {
    log.Printf("Error determining last index for key %s", key)
    return 1, err
  }
  return resp.Index, nil
}
