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
	// "golang.org/x/crypto/bcrypt"

	// etcd "github.com/coreos/etcd/clientv3"
	// "github.com/soyking/e3ch"
	// "github.com/roscopecoltran/scraper/db/redis"
	// "github.com/roscopecoltran/scraper/api"

	"github.com/jpillora/opts"
	"github.com/roscopecoltran/scraper/scraper"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/wantedly/gorm-zap"
	"go.uber.org/zap"

	// "github.com/k0kubun/pp"
	"github.com/qor/action_bar"
	"github.com/qor/help"
	"github.com/qor/media_library"
	"github.com/qor/qor"
	"github.com/roscopecoltran/admin"
	// "github.com/qor/publish2"
	// "github.com/qor/validations"
)

var VERSION = "0.0.0"

type config struct {
	*scraper.Handler `type:"embedded"`
	// etcdClient       *etcd.Client `json:"-"`

	ConfigFile string `type:"arg" help:"Path to JSON configuration file" json:"config_file" yaml:"config_file" toml:"config_file"`
	Host       string `default:"0.0.0.0" help:"Listening interface" json:"host" yaml:"host" toml:"host"`
	Port       int    `default:"8092" help:"Listening port" json:"port" yaml:"port" toml:"port"`
	NoLog      bool   `default:"false" help:"Disable access logs" json:"logs" yaml:"logs" toml:"logs"`
	EtcdHost   string `default:"etcd-1,etcd-2" help:"Listening interface" json:"etcd_host" yaml:"etcd_host" toml:"etcd_host"`
	EtcdPort   int    `default:"2379" help:"Listening port" json:"etcd_port" yaml:"etcd_port" toml:"etcd_port"`

	RedisAddr string `default:"127.0.0.1:6379" help:"Redis Addr" json:"redis_addr" yaml:"redis_addr" toml:"redis_addr"`
	RedisHost string `default:"127.0.0.1" help:"Redis host" json:"redis_host" yaml:"redis_host" toml:"redis_host"`
	RedisPort string `default:"6379" help:"Redis port" json:"redis_port" yaml:"redis_port" toml:"redis_port"`
	// redis.UseRedis(rhost)
}

var (
	DB        *gorm.DB
	AdminUI   *admin.Admin
	ActionBar *action_bar.ActionBar

	Tables = []interface{}{
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
		&scraper.Extractor{},
		&scraper.ExtractConfig{},
		&scraper.OpenAPIConfig{},
		&scraper.OpenAPISpecsConfig{},
	}

	logger  *zap.Logger
	errInit error
)

func main() {

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

	fmt.Printf(" - IsLogger? %t \n", h.Log)
	fmt.Printf(" - IsTruncateTables? %t \n", h.Config.Truncate)
	fmt.Printf(" - IsMigrateEndpoints? %t \n", h.Config.Migrate)

	// Register route
	mux := http.NewServeMux()

	// initEtcd()
	// redis.UseRedis(c.RedisHost)
	// scraper.ConvertToJsonSchema()

	if h.Config.Dashboard {
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

		scraper.MigrateTables(DB, h.Config.Truncate, Tables...)
		initDashboard()
		// amount to /admin, so visit `/admin` to view the admin interface
		AdminUI.MountTo("/admin", mux)
	}

	// if h.Config.IsApi {
	// 	api.API.MountTo("/api", mux)
	// }

	mux.Handle("/", h)
	if h.Config.Migrate {
		scraper.MigrateEndpoints(DB, h.Config)
	}

	log.Printf("Listening on: %s:%d", c.Host, c.Port)
	log.Fatal(http.ListenAndServe(c.Host+":"+strconv.Itoa(c.Port), mux))

}

func initDashboard() {

	// Initalize
	// AdminUI = admin.New(&qor.Config{DB: db.DB.Set(publish2.VisibleMode, publish2.ModeOff).Set(publish2.ScheduleMode, publish2.ModeOff)})
	AdminUI = admin.New(&qor.Config{DB: DB})

	// Meta info
	AdminUI.SetSiteName("Sniperkit-Scraper Config")

	// Auth
	// AdminUI.SetAuth(auth.AdminAuth{})

	// Assets FileSystem
	// AdminUI.SetAssetFS(bindatafs.AssetFS)

	// Menu(s)
	AdminUI.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"}) // // Add Dashboard

	// Categories (Scrapers, Providers)
	// topic := AdminUI.AddResource(&scraper.Topic{}) //, &admin.Config{Menu: []string{"Source Management"}})
	// topic.Meta(&admin.Meta{Name: "Topics", Type: "select_many"})

	// category := Admin.AddResource(&models.Category{}, &admin.Config{Menu: []string{"Product Management"}, Priority: -3})
	// category.Meta(&admin.Meta{Name: "Categories", Type: "select_many"})

	// Activity
	queries := AdminUI.AddResource(&scraper.Query{}, &admin.Config{Menu: []string{"Activity"}})
	//query := queries.Meta(&admin.Meta{Name: "Keywords"}).Resource
	queries.NewAttrs(&admin.Section{
		Rows: [][]string{{"InputQuery", "Blocked"}},
	})
	queries.EditAttrs(&admin.Section{
		Rows: [][]string{{"InputQuery", "Blocked"}},
	})

	// Groups of Scrapers
	group := AdminUI.AddResource(&scraper.Group{}, &admin.Config{Menu: []string{"Classify Data"}})
	// group.Meta(&admin.Meta{Name: "Groups", Type: "select_many"})

	// Add Asset Manager, for rich editor
	assetManager := AdminUI.AddResource(&media_library.AssetManager{}, &admin.Config{Invisible: true})

	// Add Help
	Help := AdminUI.NewResource(&help.QorHelpEntry{}, &admin.Config{Menu: []string{"Help"}})
	Help.GetMeta("Body").Config = &admin.RichEditorConfig{AssetManager: assetManager}

	// Providers
	provider := AdminUI.AddResource(&scraper.Provider{}, &admin.Config{Menu: []string{"Classify Data"}})
	providerWebRank := provider.Meta(&admin.Meta{Name: "Ranks"}).Resource
	providerWebRank.ShowAttrs("Engine", "Score")

	details := AdminUI.AddResource(&scraper.MatcherConfig{}, &admin.Config{Invisible: true})
	details.Meta(&admin.Meta{Name: "Target", Config: &admin.SelectOneConfig{Collection: scraper.TargetTypes, AllowBlank: false}})

	AdminUI.AddResource(&scraper.SelectorConfig{}, &admin.Config{Invisible: true})
	AdminUI.AddResource(&scraper.HeaderConfig{}, &admin.Config{Invisible: true})

	// Endpoints
	endpoint := AdminUI.AddResource(&scraper.Endpoint{}, &admin.Config{Menu: []string{"Web Scrapers"}})
	endpoint.Meta(&admin.Meta{Name: "Selector", Config: &admin.SelectOneConfig{Collection: scraper.SelectorEngines, AllowBlank: false}})
	endpoint.Meta(&admin.Meta{Name: "Method", Config: &admin.SelectOneConfig{Collection: scraper.MethodTypes, AllowBlank: false}})
	endpoint.Meta(&admin.Meta{Name: "Groups", Config: &admin.SelectManyConfig{SelectMode: "bottom_sheet"}})
	endpoint.IndexAttrs("Name", "Disabled", "Provider.Name", "Route", "Method")
	endpoint.SearchAttrs("Name", "Disabled", "Provider.Name", "Route", "Method")

	endpoint.Meta(&admin.Meta{Name: "Description", Config: &admin.RichEditorConfig{AssetManager: assetManager, Plugins: []admin.RedactorPlugin{
		{Name: "medialibrary", Source: "/admin/assets/javascripts/qor_redactor_medialibrary.js"},
		{Name: "table", Source: "/vendors/redactor_table.js"},
	},
		Settings: map[string]interface{}{
			"medialibraryUrl": "/admin/product_images",
		},
	}})

	endpoint.Filter(&admin.Filter{
		Name:   "Groups",
		Config: &admin.SelectOneConfig{RemoteDataResource: group},
	})

	endpoint.Filter(&admin.Filter{
		Name:   "Providers",
		Config: &admin.SelectOneConfig{RemoteDataResource: provider},
	})

	// product.SearchAttrs("Name", "Code", "Category.Name", "Brand.Name")

	headersEndpoint := endpoint.Meta(&admin.Meta{Name: "Headers"}).Resource
	headersEndpoint.NewAttrs(&admin.Section{
		Rows: [][]string{{"Key", "Value"}},
	})
	headersEndpoint.EditAttrs(&admin.Section{
		Rows: [][]string{{"Key", "Value"}},
	})

	blocksEndpoint := endpoint.Meta(&admin.Meta{Name: "Blocks"}).Resource
	blocksEndpoint.EditAttrs("Name", "Disabled", "Items", "Required", "Description", "Matchers", "StrictMode", "Debug")

	blocksEndpoint.Meta(&admin.Meta{Name: "Description", Config: &admin.RichEditorConfig{AssetManager: assetManager, Plugins: []admin.RedactorPlugin{
		{Name: "medialibrary", Source: "/admin/assets/javascripts/qor_redactor_medialibrary.js"},
		{Name: "table", Source: "/vendors/redactor_table.js"},
	},
		Settings: map[string]interface{}{
			"medialibraryUrl": "/admin/product_images",
		},
	}})

	detailsEndpoint := blocksEndpoint.Meta(&admin.Meta{Name: "Matchers"}).Resource
	detailsEndpoint.Meta(&admin.Meta{Name: "Target", Config: &admin.SelectOneConfig{Collection: scraper.TargetTypes, AllowBlank: false}})

	// Add ProductImage as Media Libraray
	ScreenshotsResource := AdminUI.AddResource(&scraper.Screenshot{}, &admin.Config{Menu: []string{"Activity"}, Priority: -1})
	ScreenshotsResource.Filter(&admin.Filter{
		Name:       "SelectedType",
		Label:      "Media Type",
		Operations: []string{"contains"},
		Config:     &admin.SelectOneConfig{Collection: [][]string{{"video", "Video"}, {"image", "Image"}, {"file", "File"}, {"video_link", "Video Link"}}},
	})
	ScreenshotsResource.IndexAttrs("File", "Title")

	// endpoint.ShowAttrs("Disabled", "Debug", "Name", "Route", "Method", "ExampleURL", "Selector", "BaseURL", "PatternURL", "Headers", "Blocks", "Extract", "StrictMode")
	/*
		endpoint.NewAttrs(
			&admin.Section{
				Title: "Status",
				Rows: [][]string{
					{"Disabled", "Debug"},
				},
			},
			&admin.Section{
				Title: "Info",
				Rows: [][]string{
					{"Name", "Slug", "Route", "Method", "ExampleURL"},
				},
			},
			&admin.Section{
				Title: "Params",
				Rows: [][]string{
					{"Selector", "BaseURL", "PatternURL"},
				},
			},
			&admin.Section{
				Title: "Headers",
				Rows: [][]string{
					{"Headers"},
				},
			},
			&admin.Section{
				Title: "Blocks",
				Rows: [][]string{
					{"Blocks"},
				},
			},
			&admin.Section{
				Title: "Bots",
				Rows: [][]string{
					{"Extract"},
				},
			},
		)

		endpoint.EditAttrs(
			&admin.Section{
				Title: "Status",
				Rows: [][]string{
					{"Disabled", "Debug"},
				},
			},
			&admin.Section{
				Title: "Info",
				Rows: [][]string{
					{"Name", "Route", "Method", "ExampleURL"},
				},
			},
			&admin.Section{
				Title: "Params",
				Rows: [][]string{
					{"Selector", "BaseURL", "PatternURL"},
				},
			},
			&admin.Section{
				Title: "Headers",
				Rows: [][]string{
					{"Headers"},
				},
			},
			&admin.Section{
				Title: "Blocks",
				Rows: [][]string{
					{"Blocks"},
				},
			},
			&admin.Section{
				Title: "Bots",
				Rows: [][]string{
					{"Extract"},
				},
			},
		)
	*/
	endpointPropertiesRes := endpoint.Meta(&admin.Meta{Name: "EndpointProperties"}).Resource
	endpointPropertiesRes.NewAttrs(&admin.Section{
		Rows: [][]string{{"Name", "Value"}},
	})
	endpointPropertiesRes.EditAttrs(&admin.Section{
		Rows: [][]string{{"Name", "Value"}},
	})

	openapi := AdminUI.AddResource(&scraper.OpenAPIConfig{}, &admin.Config{Menu: []string{"API Scrapers"}})

	// Search resources
	// AdminUI.AddSearchResource(topic)
	AdminUI.AddSearchResource(endpoint)
	AdminUI.AddSearchResource(group)
	AdminUI.AddSearchResource(provider)
	AdminUI.AddSearchResource(openapi)

}

/*
	Refs:
	- https://github.com/dwarvesf/delivr-admin/blob/develop/config/admin/admin.go
	- https://github.com/xinuxZ/wzz_qor/blob/master/app/controllers/application.go
	- https://github.com/chenxin0723/ilove/blob/master/config/routes/routes.go
	- https://github.com/reechou/erp/blob/master/app/controllers/home.go
	- https://github.com/reechou/erp/blob/master/app/controllers/category.go
	- https://github.com/reechou/erp/blob/master/app/models/order.go
	- https://github.com/sunwukonga/paypal-qor-admin/blob/master/config/admin/admin.go
	- https://github.com/sunwukonga/paypal-qor-admin/blob/master/config/admin/admin.go
	- https://github.com/angeldm/optiqor/blob/master/app/controllers/application.go
	- https://github.com/angeldm/optiqor/blob/master/config/admin/admin.go
	- https://github.com/xinuxZ/wzz_qor/blob/master/config/admin/admin.go
	- https://github.com/sunwukonga/qor-scbn/blob/devmaster/config/admin/admin.go
	- https://github.com/damonchen/beezhu/blob/master/config/admin/admin.go
	- https://github.com/sunfmin/beego_with_qor/blob/master/main.go (beego+qor)

	- https://github.com/yalay/picCms/blob/dl/models/download.go
	- https://github.com/yalay/picCms/blob/dl/models/lang.go
	- https://github.com/8legd/hugocms/blob/master/qor/models/release.go
	- https://github.com/ROOT005/managesys/blob/master/models/client.go#L41
	- https://github.com/enlivengo/admincore/tree/master/views

	- https://github.com/ROOT005/com_web/blob/master/models/project.go
	- https://github.com/ROOT005/com_web/blob/master/models/website.go

*/
