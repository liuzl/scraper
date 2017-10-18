package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/roscopecoltran/admin"
	"github.com/roscopecoltran/scraper/scraper"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/wantedly/gorm-zap"
	"go.uber.org/zap"

	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"

	"github.com/jpillora/opts"
	"github.com/k0kubun/pp"
	// "github.com/roscopecoltran/e3ch"
	// "github.com/roscopecoltran/e3w/routers"
	// "github.com/roscopecoltran/e3w/conf"
	// "github.com/roscopecoltran/e3w/e3ch"
	// "github.com/gin-contrib/cache"
	// "github.com/aviddiviner/gin-limit"
	// "github.com/gin-gonic/contrib/cache"
	// "github.com/gin-gonic/contrib/secure"
	// "github.com/gin-gonic/contrib/static"
	// "github.com/ashwanthkumar/slack-go-webhook"
	// "github.com/carlescere/scheduler"
	// "github.com/jungju/qor_admin_auth"
	// "github.com/qor/publish2"
	// "github.com/qor/validations"
	// "golang.org/x/crypto/bcrypt"
	// "github.com/roscopecoltran/scraper/db/redis"
	// "github.com/roscopecoltran/scraper/api"
)

var VERSION = "0.0.0"

type config struct {
	*scraper.Handler `type:"embedded"`

	ConfigFile string `type:"arg" help:"Path to JSON configuration file" json:"config_file" yaml:"config_file" toml:"config_file"`
	Host       string `default:"0.0.0.0" help:"Listening interface" json:"host" yaml:"host" toml:"host"`
	Port       int    `default:"8092" help:"Listening port" json:"port" yaml:"port" toml:"port"`
	NoLog      bool   `default:"false" help:"Disable access logs" json:"logs" yaml:"logs" toml:"logs"`

	EtcdHost string `default:"etcd-1,etcd-2" help:"Listening interface" json:"etcd_host" yaml:"etcd_host" toml:"etcd_host"`
	EtcdPort int    `default:"2379" help:"Listening port" json:"etcd_port" yaml:"etcd_port" toml:"etcd_port"`

	RedisAddr string `default:"127.0.0.1:6379" help:"Redis Addr" json:"redis_addr" yaml:"redis_addr" toml:"redis_addr"`
	RedisHost string `default:"127.0.0.1" help:"Redis host" json:"redis_host" yaml:"redis_host" toml:"redis_host"`
	RedisPort string `default:"6379" help:"Redis port" json:"redis_port" yaml:"redis_port" toml:"redis_port"`
	// redis.UseRedis(rhost)
}

var (
	// Serialize all modifications through these
	// commands chan interface{}
	// errors   chan error

	// Clients
	// clients     []sockjs.Session
	// clientsLock sync.RWMutex

	AdminUI *admin.Admin

	DB *gorm.DB
	// Index     bleve.Index
	// handler http.Handler
	// indexLock sync.RWMutex
	// Signalled to exit everything
	// finish chan struct{}

	// Used to control when things are done
	// wg sync.WaitGroup

	Tables = []interface{}{

		//&scraper.Endpoint{},
		&scraper.Connection{},
		&scraper.Request{},
		&scraper.Response{},

		&scraper.Screenshot{},
		&scraper.Matcher{},
		&scraper.Queries{},
		&scraper.ProviderWebRankConfig{},
		&scraper.MatcherConfig{},
		&scraper.TargetConfig{},
		&scraper.Provider{},
		&scraper.Group{},
		&scraper.Topic{},
		&scraper.Endpoint{},
		&scraper.SelectorType{},
		&scraper.ExtractorsConfig{},
		&scraper.BlocksConfig{},
		&scraper.HeaderConfig{},
		&scraper.SelectorConfig{},
		&scraper.ExtractConfig{},
		&scraper.OpenAPIConfig{},
		&scraper.OpenAPISpecsConfig{},
	}

	logger  *zap.Logger
	errInit error
)

func main() {
	useGinWrap := false
	logger, errInit = zap.NewProduction()

	h := &scraper.Handler{Log: true}
	c := config{
		Handler: h,
		Host:    "0.0.0.0",
		Port:    3000,
	}

	opts.New(&c).
		Repo("github.com/roscopecoltran/scraper").
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

	var cerr error
	fmt.Printf("Etcd.Disabled? %t \n", h.Etcd.Disabled)
	fmt.Printf("Etcd.InitCheck? %t \n", h.Etcd.InitCheck)
	fmt.Printf("Etcd.Debug? %t \n", h.Etcd.Debug)
	c.Etcd.E3ch, cerr = c.Etcd.NewE3chClient()
	if cerr != nil {
		fmt.Println("Could not connect to the ETCD cluster, error: ", cerr)
	}

	mux := http.NewServeMux() // Register route

	if h.Config.Debug {
		fmt.Printf(" - IsLogger? %t \n", h.Log)
		fmt.Println(" - Env params: ")
		pp.Println(h.Config.Env.VariablesTree)
	}

	if h.Config.Dashboard {
		if h.Config.Debug {
			fmt.Printf(" - IsTruncateTables? %t \n", h.Config.Truncate)
			fmt.Printf(" - IsMigrateEndpoints? %t \n", h.Config.Migrate)
		}
		DB, errInit = gorm.Open("sqlite3", "admin.db")
		if errInit != nil {
			panic("failed to connect database")
		}
		defer DB.Close()

		if h.Config.Debug {
			DB.LogMode(true)
			if errInit == nil {
				DB.SetLogger(gormzap.New(logger))
			}
		}

		scraper.MigrateTables(DB, h.Config.Truncate, Tables...) // Create RDB datastore
		initDashboard()
		AdminUI.MountTo("/admin", mux) // amount to /admin, so visit `/admin` to view the admin interface

	}

	// Experimental
	// redis.UseRedis(c.RedisHost)
	// scraper.ConvertToJsonSchema()
	// scraper.SeedAlexaTop1M()

	// if h.Config.IsApi {
	// 	api.API.MountTo("/api", mux)
	// }

	mux.Handle("/", h)
	if h.Config.Migrate {
		scraper.MigrateEndpoints(DB, h.Config)
	}

	// https://github.com/dwarvesf/delivr-admin/blob/develop/config/api/api.go
	// https://github.com/dwarvesf/delivr-admin/blob/develop/main.go

	if useGinWrap { // With GIN

		r := gin.Default()

		store := persistence.NewInMemoryStore(60 * time.Second)
		if h.Config.Debug {
			pp.Println(store)
		}

		/*
			client, err := e3ch.NewE3chClient(config)
			if err != nil {
				panic(err)
			}
		*/

		// routers.InitRouters(r, config, client)

		r.Any("/*w", gin.WrapH(mux))
		if err := r.Run(fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
			log.Fatalf("Can not run server, error: %s", err)
		}
	} else {
		log.Printf("Listening on: %s:%d", c.Host, c.Port)
		log.Fatal(http.ListenAndServe(c.Host+":"+strconv.Itoa(c.Port), mux))
	}

}

/*

// import "github.com/lhside/chrome-go"
func chromeBridge() {
	// Read message from standard input.
	msg, err := chrome.Receive(os.Stdin)
	// Post message to standard output
	err := chrome.Post(msg, os.Stdout)
}

// import "github.com/sauyon/go-chromemessage/chromemsg"
func chromeBridge2() {
	msg := chromemsg.New()
	msg.Read()
	msg.Write()
}
*/
