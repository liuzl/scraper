package scraper

import (
	"fmt"
	"regexp"
	// "github.com/ksinica/flatstruct"
	// "github.com/spf13/cast"
)

/*
	Refs:
	- https://github.com/Financial-Times/vulcan-config-builder/blob/master/main.go
	- https://github.com/ruprict/vulcand-atd-transformer/blob/master/templates.go
	- https://github.com/vmattos/apps-registrator/blob/master/etcd/etcd.go
	- https://github.com/vulcand/vulcand/blob/master/engine/etcdv3ng/etcd.go
	- https://github.com/vulcand/vulcand/blob/master/engine/etcdv2ng/etcd.go
*/

var (
	// etcd tree
	numDirs, numKeys int
	// scraper config
	routeIdRegex = regexp.MustCompile("/routes/([^/]+)(?:/route)?$")
	headerRegex  = regexp.MustCompile("/routes/([^/]+)/headers/([^/]+)$")
	blockRegex   = regexp.MustCompile("/routes/([^/]+)/blocks/([^/]+)$")
	// reverse proxy / load balancer
	serverRegex     = regexp.MustCompile("/backends/([^/]+)/servers/([^/]+)$")
	frontendIdRegex = regexp.MustCompile("/frontends/([^/]+)(?:/frontend)?$")
	backendIdRegex  = regexp.MustCompile("/backends/([^/]+)(?:/backend)?$")
	hostnameRegex   = regexp.MustCompile("/hosts/([^/]+)(?:/host)?$")
	listenerIdRegex = regexp.MustCompile("/listeners/([^/]+)")
	middlewareRegex = regexp.MustCompile("/frontends/([^/]+)/middlewares/([^/]+)$")
)

type EtcdService struct {
	Disabled          bool              `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Name              string            `etcd:"name" json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	HasHealthCheck    bool              `etcd:"has_health_check" json:"has_health_check,omitempty" yaml:"has_health_check,omitempty" toml:"has_health_check,omitempty"`
	Addresses         map[string]string `etcd:"addresses" json:"addresses,omitempty" yaml:"addresses,omitempty" toml:"addresses,omitempty"`
	PathPrefixes      map[string]string `etcd:"path_prefixes" json:"path_prefixes,omitempty" yaml:"path_prefixes,omitempty" toml:"path_prefixes,omitempty"`
	PathHosts         map[string]string `etcd:"path_hosts" json:"path_hosts,omitempty" yaml:"path_hosts,omitempty" toml:"path_hosts,omitempty"`
	FailoverPredicate string            `etcd:"failover_predicate" json:"failover_predicate,omitempty" yaml:"failover_predicate,omitempty" toml:"failover_predicate,omitempty"`
	Debug             bool              `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

type EtcdRoute struct {
	Disabled bool `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	// server - endpoint handler
	Source string `etcd:"source" json:"provider,omitempty" yaml:"provider,omitempty" toml:"provider,omitempty"`
	Route  string `etcd:"router" json:"route,omitempty" yaml:"route,omitempty" toml:"route,omitempty"`
	Method string `etcd:"method" json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`
	// remote content - extraction rules and blocks
	BaseURL          string            `required:"true" etcd:"base_url" json:"base_url,omitempty" yaml:"base_url,omitempty" toml:"base_url,omitempty"`
	PatternURL       string            `required:"true" etcd:"pattern_url" json:"pattern_url" yaml:"pattern_url" toml:"pattern_url"`
	Selector         string            `etcd:"selector" default:"css" json:"selector,omitempty" yaml:"selector,omitempty" toml:"selector,omitempty"`
	HeadersIntercept map[string]string `etcd:"resp_headers_intercept" json:"resp_headers_intercept,omitempty" yaml:"resp_headers_intercept,omitempty" toml:"resp_headers_intercept,omitempty"`
	Headers          map[string]string `etcd:"headers" json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
	Blocks           map[string]string `etcd:"blocks" json:"blocks,omitempty" yaml:"blocks,omitempty" toml:"blocks,omitempty"`
	Extract          map[string]string `etcd:"extract" json:"extract,omitempty" yaml:"extract,omitempty" toml:"extract,omitempty"`
	Groups           string            `etcd:"groups" json:"groups,omitempty" yaml:"groups,omitempty" toml:"groups,omitempty"`
	StrictMode       bool              `etcd:"strict_mode" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
	Debug            bool              `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

type EtcdProxy struct {
	Disabled  bool                    `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Frontends map[string]EtcdFrontend `etcd:"frontends" json:"frontends,omitempty" yaml:"frontends,omitempty" toml:"frontends,omitempty"`
	Backends  map[string]EtcdBackend  `etcd:"backends" json:"backends,omitempty" yaml:"backends,omitempty" toml:"backends,omitempty"`
	// Middlewares map[string]EtcdMiddleware `etcd:"middlewares" json:"middlewares,omitempty" yaml:"middlewares,omitempty" toml:"middlewares,omitempty"`
	// Plugins     map[string]EtcdPlugin     `etcd:"plugins" json:"plugins,omitempty" yaml:"plugins,omitempty" toml:"plugins,omitempty"`
	Debug bool `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

type EtcdFrontend struct {
	Disabled          bool        `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	BackendID         string      `etcd:"backend_id" json:"backend_id,omitempty" yaml:"backend_id,omitempty" toml:"backend_id,omitempty"`
	Route             string      `etcd:"route" json:"route,omitempty" yaml:"route,omitempty" toml:"route,omitempty"`
	Type              string      `etcd:"type" json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Rewrite           EtcdRewrite `etcd:"rewrite" json:"rewrite,omitempty" yaml:"rewrite,omitempty" toml:"rewrite,omitempty"`
	FailoverPredicate string      `etcd:"failover_predicate" json:"failover_predicate,omitempty" yaml:"failover_predicate,omitempty" toml:"failover_predicate,omitempty"`
	Debug             bool        `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

type EtcdRewrite struct {
	Disabled   bool          `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	ID         string        `etcd:"id" json:"id,omitempty" yaml:"id,omitempty" toml:"id,omitempty"`
	Type       string        `etcd:"type" json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Priority   int           `etcd:"priority" json:"priority,omitempty" yaml:"priority,omitempty" toml:"priority,omitempty"`
	Middleware EtcdRewriteMw `etcd:"middleware" json:"middleware,omitempty" yaml:"middleware,omitempty" toml:"middleware,omitempty"`
	Debug      bool          `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

type EtcdRewriteMw struct {
	Disabled    bool   `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Regexp      string `etcd:"regexp" json:"regexp,omitempty" yaml:"regexp,omitempty" toml:"regexp,omitempty"`
	Replacement string `etcd:"replacement" json:"replacement,omitempty" yaml:"replacement,omitempty" toml:"replacement,omitempty"`
	Debug       bool   `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

type EtcdBackend struct {
	Disabled bool                  `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Servers  map[string]EtcdServer `etcd:"servers" json:"servers,omitempty" yaml:"servers,omitempty" toml:"servers,omitempty"`
	Debug    bool                  `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

type EtcdServer struct {
	Disabled bool   `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	URL      string `etcd:"url" json:"url,omitempty" yaml:"url,omitempty" toml:"url,omitempty"`
	Debug    bool   `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

type EtcdHandler struct {
	Disabled   bool                    `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Backends   map[string]EtcdBackend  `etcd:"backends" json:"backends,omitempty" yaml:"backends,omitempty" toml:"backends,omitempty"`
	Frontends  map[string]EtcdFrontend `etcd:"frontends" json:"frontends,omitempty" yaml:"frontends,omitempty" toml:"frontends,omitempty"`
	Routes     map[string]EtcdRoute    `etcd:"routes" json:"routes,omitempty" yaml:"routes,omitempty" toml:"routes,omitempty"`
	Services   map[string]EtcdService  `etcd:"services" json:"services,omitempty" yaml:"services,omitempty" toml:"services,omitempty"`
	StrictMode bool                    `etcd:"strict_mode" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
	Debug      bool                    `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

func (rc *EtcdHandler) NewEndpoint(path string, conf Endpoint) (bool, error) {
	fmt.Printf("Register new endpoint: %s \n", path)
	return true, nil
}

/*
func (eh *EtcdHandler) FlattenHandlerConfig(h Handler) (EtcdHandler.Routes, error) {
	fs := flatstruct.NewFlatStruct()
	fs.PathSeparator = "/"
	for _, e := range h.Config.Routes {
		err := c.Etcd.RecursiveCreateDir(fmt.Sprintf("/%s", e.Route)) // Create all dirs recursively...
		if err != nil {
			fmt.Println("error: ", err)
		}
		f, err := fs.Flatten(e)
		if err != nil {
			fmt.Println("error: ", err)
		}
		for k, _ := range f {
			keyParts := strings.Split(fmt.Sprintf("/%s", k), "/")
			if len(keyParts) > 1 {
				dir := len(keyParts) - 2
				key := len(keyParts) - 1
				fmt.Printf("dir='/%s%s', key='%s', cast='%s', val='%v' \n", e.Route, strings.Join(keyParts[:dir], "/"), keyParts[key], cast.ToString(f[k].Value), f[k].Value)
				err := c.Etcd.RecursiveCreateDir(fmt.Sprintf("/%s%s", e.Route, strings.Join(keyParts[:dir], "/"))) // Create all dirs recursively...
				if err != nil {
					fmt.Println("error: ", err)
				}
				err = c.Etcd.E3ch.Create(fmt.Sprintf("/%s%s/%s", e.Route, strings.Join(keyParts[:dir], "/"), keyParts[key]), cast.ToString(f[k].Value))
				if err != nil {
					fmt.Println("error: ", err)
					err = c.Etcd.E3ch.Put(fmt.Sprintf("/%s%s/%s", e.Route, strings.Join(keyParts[:dir], "/"), keyParts[key]), cast.ToString(f[k].Value))
					if err != nil {
						fmt.Println("error: ", err)
					}
				}
			}
		}
	}
}
*/
