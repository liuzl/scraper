package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	// "github.com/roscopecoltran/configor"
	// "github.com/k0kubun/pp"
)

type Handler struct {
	Config  Config            `opts:"-" json:"config,omitempty" yaml:"config,omitempty" toml:"config,omitempty"`
	Headers map[string]string `opts:"-" json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
	Auth    string            `help:"Basic auth credentials <user>:<pass>" json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	Log     bool              `opts:"-" json:"log,omitempty" yaml:"log,omitempty" toml:"log,omitempty"`
	Debug   bool              `help:"Enable debug output" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

func (h *Handler) LoadConfigFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return h.LoadConfig(b)
}

var Endpoints struct {
	Disabled bool
	Routes   []string
}

func (h *Handler) LoadConfig(b []byte) error {
	c := Config{}

	if err := json.Unmarshal(b, &c); err != nil { //json unmarshal performs selector validation
		return err
	}

	if h.Log {
		for k, e := range c.Routes {
			// pp.Print(c.Routes)
			if strings.HasPrefix(e.Route, "/") {
				e.Route = strings.TrimPrefix(e.Route, "/")
				c.Routes[k] = e
			}
			logf("Loaded endpoint: /%s", e.Route)
			Endpoints.Routes = append(Endpoints.Routes, e.Route)
			// Copy the Debug attribute
			e.Debug = h.Debug
			// Copy the Header attributes (only if they are not yet set)
			if e.Headers == nil {
				e.Headers = h.Headers
			} else {
				for k, v := range h.Headers {
					if _, ok := e.Headers[k]; !ok {
						e.Headers[k] = v
					}
				}
			}
		}
	}
	if h.Debug {
		logf("Enabled debug mode")
	}

	// pp.Print(Endpoints)

	//replace config
	h.Config = c
	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// basic auth
	if h.Auth != "" {
		u, p, _ := r.BasicAuth()
		if h.Auth != u+":"+p {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Access Denied"))
			return
		}
	}
	// always JSON!
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// admin actions
	if r.URL.Path == "" || r.URL.Path == "/" {
		get := false
		if r.Method == "GET" {
			get = true
		} else if r.Method == "POST" {
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(jsonerr(err))
				return
			}
			if err := h.LoadConfig(b); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(jsonerr(err))
				return
			}
			get = true
		}
		if !get {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write(jsonerr(errors.New("Use GET or POST")))
			return
		}
		b, _ := json.MarshalIndent(h.Config, "", "  ")
		w.Write(b)
		return
	}
	// endpoint id (excludes root slash)
	id := r.URL.Path[1:]
	// load endpoint
	endpoint := h.Endpoint(id)
	if endpoint == nil {
		w.WriteHeader(404)
		w.Write(jsonerr(fmt.Errorf("Endpoint /%s not found", id)))
		return
	}
	// convert url.Values into map[string]string
	values := map[string]string{}
	for k, v := range r.URL.Query() {
		values[k] = v[0]
	}
	// execute query
	res, err := endpoint.Execute(values)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonerr(err))
		return
	}
	// encode as JSON
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		w.Write([]byte("JSON Error: " + err.Error()))
	}
}

// Endpoint will return the Handler's Endpoint from its Config
func (h *Handler) Endpoint(path string) *Endpoint {
	var keyCfg int
	for k, v := range h.Config.Routes {
		if v.Route == path {
			keyCfg = k
			break
		}
	}
	if h.Config.Routes[keyCfg] != nil {
		return h.Config.Routes[keyCfg]
	}
	return nil
}
