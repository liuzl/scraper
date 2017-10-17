package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	// "github.com/roscopecoltran/configor"
)

type Handler struct {
	Disabled bool `default:"false" help:"Disable handler init" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`

	Env EnvConfig `opts:"-" json:"env" yaml:"env" file:"environment" toml:"env"`

	Config  Config            `opts:"-" json:"config,omitempty" yaml:"config,omitempty" toml:"config,omitempty"`
	Headers map[string]string `opts:"-" json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`

	Auth string `help:"Basic auth credentials <user>:<pass>" json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	Log  bool   `default:"false" opts:"-" json:"log,omitempty" yaml:"log,omitempty" toml:"log,omitempty"`

	Debug bool `default:"false" help:"Enable debug output" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
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

	if len(c.Env.Files) > 0 {
		envVars, err := godotenv.Read(c.Env.Files...)
		if err != nil {
			return err
		}
		c.Env.VariablesList = envVars
		envVarsTree := make(map[string]map[string]string)
		for k, v := range envVars {
			var varParentKey, varChildrenKey string
			varParts := strings.Split(k, "_")
			if len(varParts) > 1 {
				varParentKey = varParts[0]
				varChildrenKey = strings.Join(varParts[1:], "_")
			}
			if v != "" && varParentKey != "" && varChildrenKey != "" {
				envVarsTree[varParentKey] = make(map[string]string)
				envVarsTree[varParentKey][varChildrenKey] = v
			}
		}
		c.Env.VariablesTree = envVarsTree
	}

	if h.Log {
		for k, e := range c.Routes {
			if strings.HasPrefix(e.Route, "/") {
				e.Route = strings.TrimPrefix(e.Route, "/")
				c.Routes[k] = e
			}
			logf("Loaded endpoint: /%s", e.Route)
			Endpoints.Routes = append(Endpoints.Routes, e.Route)
			e.Debug = h.Debug                  // Copy the Debug attribute
			if len(h.Headers) > 0 && h.Debug { // Copy the Header attributes (only if they are not yet set)
				fmt.Printf("h.Headers, len=%d:\n", len(h.Headers))
				pp.Println(h.Headers)
			}
			for k, v := range e.HeadersJSON {
				if len(e.HeadersJSON) > 0 && h.Debug {
					pp.Println("header key: ", k)
					pp.Println("header val: ", v)
				}
				for kl, vl := range c.Env.VariablesList {
					holderKey := fmt.Sprintf("{{%s}}", strings.Replace(kl, "\"", "", -1))
					v = strings.Replace(v, holderKey, vl, -1)
				}
				e.HeadersJSON[k] = strings.Trim(v, " ")
			}
			if e.HeadersJSON == nil {
				e.HeadersJSON = h.Headers
			} else {
				for k, v := range h.Headers {
					if _, ok := e.HeadersJSON[k]; !ok {
						e.HeadersJSON[k] = v
					}
				}
			}
			if len(e.HeadersJSON) > 0 && h.Debug {
				fmt.Printf("e.HeadersJSON, len=%d:\n", len(e.HeadersJSON))
				pp.Println(e.HeadersJSON)
			}
		}
	}
	if h.Debug {
		logf("Enabled debug mode")
	}
	h.Config = c //replace config
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
