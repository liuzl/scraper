package main

import (
	/*"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"io/ioutil"
	*/
	"errors"
	"sync"

	etcd "github.com/coreos/etcd/clientv3"
	// clientbuilder "github.com/tonnerre/go-etcd-clientbuilder"
)

var etcdClient *etcd.Client
var etcdOnce *sync.Once
var etcdOnceError error

/*
	Refs:
	- https://github.com/tonnerre/go-etcd-clientbuilder/blob/master/autoconf/autoconf.go
	-
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
