package main

import (
	"errors"
	"fmt"
	"sync"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/soyking/e3ch"
)

var (
	// Etcd v3.x - Flags
	EtcdHost string `default:"etcd-1,etcd-2" help:"Listening interface" json:"etcd_host" yaml:"etcd_host" toml:"etcd_host"`
	EtcdPort int    `default:"2379" help:"Listening port" json:"etcd_port" yaml:"etcd_port" toml:"etcd_port"`
	// Etcd v3.x - Objects
	etcdClient    *etcd.Client
	etcdOnce      *sync.Once
	etcdOnceError error
)

/*
	Refs:
	- https://github.com/tonnerre/go-etcd-clientbuilder/blob/master/autoconf/autoconf.go
	-
*/

func initEtcd() {

	fmt.Printf("Connecting to ETCD v3.x cluster...\n")

	// initial etcd v3 client
	// strings.Split(*etcdServer, ",")
	e3Clt, err := etcd.New(etcd.Config{
		Endpoints:   []string{"etcd1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	// pp.Print(e3Clt)

	// new e3ch client with namespace(rootKey)
	clt, err := client.New(e3Clt, "scraper")
	if err != nil {
		panic(err)
	}
	// pp.Print(clt)

	// set the rootKey as directory
	err = clt.FormatRootKey()
	if err != nil {
		panic(err)
	}

	clt.CreateDir("/dir1")
	clt.Create("/dir1/key1", "")
	clt.Create("/dir", "")
	clt.Put("/dir1/key1", "value1")
	clt.Get("/dir1/key1")
	clt.List("/dir1")
	clt.Delete("/dir")

	clt.List("/")

}

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