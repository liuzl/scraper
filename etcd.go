package main

import (
	"fmt"
	"sync"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/roscopecoltran/e3ch"
)

var (
	etcdClient    *etcd.Client
	etcdOnce      *sync.Once
	etcdOnceError error
)

func initEtcd() (*client.EtcdHRCHYClient, error) {

	fmt.Printf("Connecting to ETCD v3.x cluster...\n")

	// initial etcd v3 client
	// strings.Split(*etcdServer, ",")
	e3Clt, err := etcd.New(etcd.Config{
		Endpoints:   []string{"etcd1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("failed to init initial etcd v3 client, error: ", err)
		return nil, err
	}
	// pp.Print(e3Clt)

	// new e3ch client with namespace(rootKey)
	clt, err := client.New(e3Clt, "scraper")
	if err != nil {
		fmt.Println("failed to init e3ch client with namespace(rootKey), error: ", err)
		return nil, err
	}
	// pp.Print(clt)

	// set the rootKey as directory
	err = clt.FormatRootKey()
	if err != nil {
		fmt.Println("failed to  set the rootKey as directory, error: ", err)
		return nil, err
	}

	// Quick Test
	clt.CreateDir("/dir1")
	clt.Create("/dir1/key1", "")
	clt.Create("/dir", "")
	clt.Put("/dir1/key1", "value1")
	clt.Get("/dir1/key1")
	clt.List("/dir1")
	clt.Delete("/dir")
	clt.List("/")

	return clt, nil
}

/*
	Refs:
	- https://github.com/tonnerre/go-etcd-clientbuilder/blob/master/autoconf/autoconf.go
*/

/*
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resp, err := cli.Put(ctx, "sample_key", "sample_value")
	cancel()
	if err != nil {
	    // handle error!
	}
*/

/*
	resp, err := cli.Put(ctx, "", "")
	if err != nil {
		switch err {
		case context.Canceled:
			log.Fatalf("ctx is canceled by another routine: %v", err)
		case context.DeadlineExceeded:
			log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
		case rpctypes.ErrEmptyKey:
			log.Fatalf("client-side error: %v", err)
		default:
			log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
		}
	}
*/
