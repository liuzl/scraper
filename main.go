package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/go-fsnotify/fsnotify"
	"github.com/googollee/go-socket.io"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jpillora/opts"
	"github.com/k0kubun/pp"
	"github.com/roscopecoltran/admin"
	"github.com/roscopecoltran/scraper/scraper"
	"github.com/wantedly/gorm-zap"
	"go.uber.org/zap"
	"gopkg.in/olahol/melody.v1"
	// fsnotify "gopkg.in/fsnotify.v1"
	// "github.com/valyala/fasthttp"
	// "github.com/geekypanda/httpcache"
	// "github.com/meission/router"
	// "github.com/go-zoo/bone"
	// "github.com/birkelund/boltdbcache"
	// "golang.org/x/oauth2"
	// "github.com/mickep76/flatten"
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
	// https://github.com/aliostad/deep-learning-lang-detection/blob/1180fba0d2a7f6b470cb3c9a363b560787f5e7c5/data/test/go/ec5f82a852d053a084edbc39ac4b56f9381b7cf9test.go
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
	// Index     bleve.Index

	// handler http.Handler
	// indexLock sync.RWMutex
	// Signalled to exit everything
	// finish chan struct{}

	// Used to control when things are done
	// wg sync.WaitGroup

	AdminUI *admin.Admin
	DB      *gorm.DB

	Tables = []interface{}{
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

var cacheDuration = 3600 * time.Second

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

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
			//if err := h.LoadConfigorFile(c.ConfigFile); err != nil {
			if err := h.LoadConfigFile(c.ConfigFile); err != nil {
				log.Printf("[scraper] Failed to load configuration: %s", err)
			} else {
				log.Printf("[scraper] Successfully loaded new configuration")
			}
		}
	}()

	m := melody.New()
	w, _ := fsnotify.NewWatcher()
	go func() {
		for {
			ev := <-w.Events
			if ev.Op == fsnotify.Write {
				content, _ := ioutil.ReadFile(ev.Name)
				fmt.Println("ev.Name:", ev.Name)
				fmt.Printf("content: %s\n", content)
				m.Broadcast(content)

				if err := h.LoadConfigFile(c.ConfigFile); err != nil {
					log.Printf("[scraper] Failed to load configuration: %s", err)
				} else {
					log.Printf("[scraper] Successfully loaded new configuration")
				}

			}
		}
	}()

	//if err := h.LoadConfigorFile(c.ConfigFile); err != nil {
	if err := h.LoadConfigFile(c.ConfigFile); err != nil {
		log.Fatal(err)
	}

	m.HandleConnect(func(s *melody.Session) {
		content, _ := ioutil.ReadFile(c.ConfigFile)
		s.Write(content)
		fmt.Println("c.ConfigFile:", c.ConfigFile)
		fmt.Printf("content: %s\n", content)
	})
	w.Add(c.ConfigFile)

	// cache
	// https://github.com/garycarr/httpcache/blob/f039dd6ff44cf40d52e8e86ef10bff41e592fd48/README.md
	fmt.Printf("Scraper.NumCPU: %d\n", runtime.NumCPU())
	fmt.Printf("Scraper.useGinWrap: %t\n", useGinWrap)
	fmt.Printf("Scraper.Etcd.Disabled? %t \n", h.Etcd.Disabled)
	fmt.Printf("Scraper.Etcd.InitCheck? %t \n", h.Etcd.InitCheck)
	fmt.Printf("Scraper.Etcd.Debug? %t \n", h.Etcd.Debug)
	e3ch, err := c.Etcd.NewE3chClient()
	if err != nil {
		fmt.Println("Could not connect to the ETCD cluster, error: ", err)
	}

	if e3ch != nil {
		h.Etcd.Handler = h
		h.Etcd.E3ch = e3ch
	}

	// if useBoneMux
	// mux := bone.New()
	mux := http.NewServeMux() // Register route

	/*
	   // mux.Get, Post, etc ... takes http.Handler
	   mux.Get("/home/:id", http.HandlerFunc(HomeHandler))
	   mux.Get("/profil/:id/:var", http.HandlerFunc(ProfilHandler))
	   mux.Post("/data", http.HandlerFunc(DataHandler))

	   // Support REGEX Route params
	   mux.Get("/index/#id^[0-9]$", http.HandlerFunc(IndexHandler))

	   // Handle take http.Handler
	   mux.Handle("/", http.HandlerFunc(RootHandler))
	*/

	if h.Config.Debug {
		fmt.Printf(" - IsLogger? %t \n", h.Log)
		fmt.Println(" - Env params: ")
		pp.Println(h.Config.Env.VariablesTree)
	}

	if h.Config.Migrate {
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
	}

	if h.Config.Dashboard {
		initDashboard()
		AdminUI.MountTo("/admin", mux) // amount to /admin, so visit `/admin` to view the admin interface
	}

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// https://github.com/googollee/go-socket.io/blob/master/example/main.go
	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
		so.Join("chat")
		so.On("chat message", func(msg string) {
			fmt.Println(so, msg)
			log.Println("emit:", so.Emit("chat message", msg))
			so.BroadcastTo("chat", "chat message", msg)
		})
		so.On("disconnection", func() {
			log.Println("on disconnect")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	// Experimental
	// redis.UseRedis(c.RedisHost)
	// scraper.ConvertToJsonSchema()
	// scraper.SeedAlexaTop1M()
	// h = scraper.NewRequestCacher(mux, "./shared/cache/scraper")
	mux.Handle("/", h)
	//	mux.HandleFunc("/ws", m.HandleRequest())
	mux.Handle("/socket.io/", server)

	mux.HandleFunc("/favicon.ico", scraper.FaviconHandler)
	mux.HandleFunc("/test", handler)

	// GetFunc, PostFunc etc ... takes http.HandlerFunc
	// mux.GetFunc("/test", Handler)

	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	if h.Config.Migrate {
		scraper.MigrateEndpoints(DB, h.Config, e3ch)
	}

	if useGinWrap { // With GIN

		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		store := persistence.NewInMemoryStore(60 * time.Second)
		if h.Config.Debug {
			fmt.Println("store: ")
			pp.Println(store)
		}

		r.Any("/*w", gin.WrapH(mux))
		if err := r.Run(fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
			log.Fatalf("Can not run server, error: %s", err)
		}

	} else {

		log.Printf("Listening on: %s:%d", c.Host, c.Port)
		log.Fatal(http.ListenAndServe(c.Host+":"+strconv.Itoa(c.Port), mux))

	}

}

// "handler" is our handler function. It has to follow the function signature of a ResponseWriter and Request type
// as the arguments.
func handler(w http.ResponseWriter, r *http.Request) {
	// For this case, we will always pipe "Hello World" into the response writer
	fmt.Fprintf(w, "Hello World!")
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
