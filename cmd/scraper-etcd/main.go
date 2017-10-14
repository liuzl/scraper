package main

// luc

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/jpillora/opts"
	// "github.com/k0kubun/pp"
	"github.com/roscopecoltran/scraper/scraper"
	"github.com/soyking/e3ch"
)

var VERSION = "0.0.0"

type config struct {
	*scraper.Handler `type:"embedded"`
	// etcdClient       *etcd.Client `json:"-"`

	ConfigFile string `type:"arg" help:"Path to JSON configuration file" json:"config_file"`
	Host       string `default:"0.0.0.0" help:"Listening interface" json:"host"`
	Port       int    `default:"8092" help:"Listening port" json:"port"`
	NoLog      bool   `default:"false" help:"Disable access logs" json:"logs"`
	EtcdHost   string `default:"etcd-1,etcd-2" help:"Listening interface" json:"etcd_host"`
	EtcdPort   int    `default:"2379" help:"Listening port" json:"etcd_port"`
}

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

func main() {

	h := &scraper.Handler{Log: true}

	c := config{
		Handler: h,
		Host:    "0.0.0.0",
		Port:    3000,
	}

	opts.New(&c).
		Repo("github.com/jpillora/scraper").
		Version(VERSION).
		Parse()

	h.Log = !c.NoLog

	go func() {
		for {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGHUP)
			<-sig
			if err := h.LoadConfigFile(c.ConfigFile); err != nil {
				log.Printf("[scraper] Failed to load configuration: %s", err)
			} else {
				log.Printf("[scraper] Successfully loaded new configuration")
			}
		}
	}()

	if err := h.LoadConfigFile(c.ConfigFile); err != nil {
		log.Fatal(err)
	}

	initEtcd()

	log.Printf("[scraper] Listening on %d...", c.Port)
	log.Fatal(http.ListenAndServe(c.Host+":"+strconv.Itoa(c.Port), h))
}
