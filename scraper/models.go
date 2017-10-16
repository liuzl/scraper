package scraper

import (
	"github.com/jinzhu/gorm"
)

// Result represents a result
type Result map[string]string

// Create a GORM-backend model
type Provider struct {
	gorm.Model
	Name string
}

// Config represents...
type Config struct {
	gorm.Model
	Disabled bool `default:"false" help:"Disable handler init" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`

	Port      int         `default:"3000" json:"port,omitempty" yaml:"port,omitempty" toml:"port,omitempty"`
	Routes    []*Endpoint `gorm:"-" json:"routes,omitempty" yaml:"routes,omitempty" toml:"routes,omitempty"`
	Dashboard bool        `default:"false" help:"Initialize the Administration Interface" json:"dashboard,omitempty" yaml:"dashboard,omitempty" toml:"dashboard,omitempty"`
	Truncate  bool        `default:"true" help:"Truncate previous data" json:"truncate,omitempty" yaml:"truncate,omitempty" toml:"truncate,omitempty"`
	Migrate   bool        `default:"true" help:"Migrate to admin dashboard" json:"migrate,omitempty" yaml:"migrate,omitempty" toml:"migrate,omitempty"`

	Debug bool `default:"false" help:"Enable debug output" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

// Endpoint represents a single remote endpoint. The performed query can be modified between each call by parameterising URL. See documentation.
type Endpoint struct {
	gorm.Model
	Disabled bool `default:"false" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`

	ProviderStr string `gorm:"-" json:"provider,omitempty" yaml:"provider,omitempty" toml:"provider,omitempty"`
	// Provider    *Provider `json:"provider_orm,omitempty" yaml:"provider_orm,omitempty" toml:"provider_orm,omitempty"`

	Route  string `json:"route,omitempty" yaml:"route,omitempty" toml:"route,omitempty"`
	Name   string `gorm:"index" json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	Method string `gorm:"index" json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`

	BaseURL    string `gorm:"index" json:"base_url,omitempty" yaml:"base_url,omitempty" toml:"base_url,omitempty"`
	PatternURL string `json:"url" yaml:"url" toml:"url"`
	ExampleURL string `json:"example_url" yaml:"example_url" toml:"example_url"`

	Body     string `json:"body,omitempty" yaml:"body,omitempty" toml:"body,omitempty"`
	Selector string `gorm:"index" default:"css" json:"selector,omitempty" yaml:"selector,omitempty" toml:"selector,omitempty"`

	HeadersJSON map[string]string         `gorm:"-" json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
	BlocksJSON  map[string]SelectorConfig `gorm:"-" json:"blocks,omitempty" yaml:"blocks,omitempty" toml:"blocks,omitempty"`

	Headers []*HeaderConfig   `json:"headers_orm,omitempty" yaml:"headers_orm,omitempty" toml:"headers_orm,omitempty"`
	Blocks  []*SelectorConfig `json:"blocks_orm,omitempty" yaml:"blocks_orm,omitempty" toml:"blocks_orm,omitempty"`

	Extract   ExtractConfig `default:"false" json:"extract,omitempty" yaml:"extract,omitempty" toml:"extract,omitempty"`
	MinFields int           `json:"-" yaml:"-" toml:"-"`
	Count     string        `gorm"-" json:"-" yaml:"-" toml:"-"`

	Debug      bool `default:"false" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	StrictMode bool `default:"false" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
}

//type HeadersProperties []HeaderConfig
//type BlocksProperties []SelectorConfig
//type DetailsProperties []ExtractorORM

// SelectorConfig represents a content selection rule for a single URL Pattern.
type SelectorConfig struct {
	gorm.Model
	EndpointID uint
	Items      string                `json:"items,omitempty" yaml:"items,omitempty" toml:"items,omitempty"`
	Details    map[string]Extractors `gorm:"-" json:"details,omitempty" yaml:"details,omitempty" toml:"details,omitempty"`
	Matchers   []*MatcherConfig      `json:"matchers,omitempty" yaml:"matchers,omitempty" toml:"matchers,omitempty"`
	StrictMode bool                  `default:"false" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
	Required   bool                  `default:"true" json:"required,omitempty" yaml:"required,omitempty" toml:"required,omitempty"`
	Slug       string                `json:"slug,omitempty" yaml:"slug,omitempty" toml:"slug,omitempty"`
	Debug      bool                  `default:"true" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

// Extractor represents a pair of css selector and extracted node.
type Extractor struct {
	val string
	fn  extractorFn `gorm:"-"`
}

// Extractor represents a pair of css selector and extracted node.
type MatcherConfig struct {
	gorm.Model
	SelectorConfigID uint
	Target           string // TargetConfig
	Matcher          string
	// fn  extractorFn
}

var TargetTypes = []string{"title", "desc", "image", "price", "stock", "count", "url", "tag", "extra", "cat"}
var TransportTypes = []string{"http", "https", "grpc", "tcp", "udp", "udp", "udp", "inproc", "ipc", "tlstcp", "ws", "wss"}
var MethodTypes = []string{"GET", "POST"}
var SelectorEngines = []string{"css", "xpath", "json", "xml", "csv"}

type SelectorType struct {
	gorm.Model
	Name   string
	Engine string
}

type TargetConfig struct {
	gorm.Model
	// EndpointID uint
	Name string
}

type HeaderConfig struct {
	gorm.Model
	EndpointID uint
	Key        string
	Value      string
}

type BlocksConfig struct {
	gorm.Model
	Key   string
	Value SelectorConfig
}

type ExtractorsConfig struct {
	gorm.Model
	Key   string
	Value Extractors
}

// ExtractConfig represents a single sub-extraction rules url content configuration.
type ExtractConfig struct {
	gorm.Model
	Debug     bool `default:"true" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	Links     bool `default:"true" json:"links,omitempty" yaml:"links,omitempty" toml:"links,omitempty"`
	Meta      bool `default:"true" json:"meta,omitempty" yaml:"meta,omitempty" toml:"meta,omitempty"`
	OpenGraph bool `default:"true" json:"opengraph,omitempty" yaml:"opengraph,omitempty" toml:"opengraph,omitempty"`
}
