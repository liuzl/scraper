package main

type Extractors []Extractor

// Result represents a result
type Result map[string]string

// Config represents...
type Config struct {
	Port   int         `default:"3000" json:"port,omitempty" yaml:"port,omitempty" toml:"port,omitempty"`
	Routes []*Endpoint `json:"routes,omitempty" yaml:"routes,omitempty" toml:"routes,omitempty"`
}

// Endpoint represents a single remote endpoint. The performed query can be modified between each call by parameterising URL. See documentation.
type Endpoint struct {
	Disabled bool `default:"false" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`

	Provider Provider

	Route      string `json:"route,omitempty" yaml:"route,omitempty" toml:"route,omitempty"`
	Name       string `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	Method     string `json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`
	BaseURL    string `json:"base_url,omitempty" yaml:"base_url,omitempty" toml:"base_url,omitempty"`
	PatternURL string `json:"url" yaml:"url" toml:"url"`
	Body       string `json:"body,omitempty" yaml:"body,omitempty" toml:"body,omitempty"`

	Selector string `default:"css" json:"selector,omitempty" yaml:"selector,omitempty" toml:"selector,omitempty"`
	// Headers  map[string]string         `json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
	// Blocks   map[string]SelectorConfig `json:"blocks,omitempty" yaml:"blocks,omitempty" toml:"blocks,omitempty"`

	Headers []HeaderConfig   `json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
	Blocks  []SelectorConfig `json:"blocks,omitempty" yaml:"blocks,omitempty" toml:"blocks,omitempty"`

	Extract   ExtractConfig `default:"false" json:"extract,omitempty" yaml:"extract,omitempty" toml:"extract,omitempty"`
	MinFields int           `json:"-" yaml:"-" toml:"-"`
	Count     string        `json:"-" yaml:"-" toml:"-"`

	Debug      bool `default:"false" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	StrictMode bool `default:"false" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
}

var MethodTypes = []string{"GET", "POST"}
var SelectorEngines = []string{"css", "xpath", "mxj", "gabs"}

type SelectorType struct {
	Name   string
	Engine string
}

type HeaderConfig struct {
	Key   string
	Value string
}

type BlocksConfig struct {
	Key   string
	Value SelectorConfig
}

// SelectorConfig represents a content selection rule for a single URL Pattern.
type SelectorConfig struct {
	Slug     string `json:"slug,omitempty" yaml:"slug,omitempty" toml:"slug,omitempty"`
	Debug    bool   `default:"true" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	Required bool   `default:"true" json:"required,omitempty" yaml:"required,omitempty" toml:"required,omitempty"`
	Selector string `default:"css" json:"selector,omitempty" yaml:"selector,omitempty" toml:"selector,omitempty"`
	Items    string `json:"items,omitempty" yaml:"items,omitempty" toml:"items,omitempty"`
	//Details    map[string]Extractors `json:"details,omitempty" yaml:"details,omitempty" toml:"details,omitempty"`
	Details    []Extractor `json:"details,omitempty" yaml:"details,omitempty" toml:"details,omitempty"`
	StrictMode bool        `default:"false" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
}

type ExtractorsConfig struct {
	Key   string
	Value Extractors
}

// Extractor represents a pair of css selector and extracted node.
type Extractor struct {
	Key  string
	Node string
	// fn  extractorFn
}

// ExtractConfig represents a single sub-extraction rules url content configuration.
type ExtractConfig struct {
	Debug     bool `default:"true" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	Links     bool `default:"true" json:"links,omitempty" yaml:"links,omitempty" toml:"links,omitempty"`
	Meta      bool `default:"true" json:"meta,omitempty" yaml:"meta,omitempty" toml:"meta,omitempty"`
	OpenGraph bool `default:"true" json:"opengraph,omitempty" yaml:"opengraph,omitempty" toml:"opengraph,omitempty"`
}

type Entity struct {
	Id          string  `json:"id"`
	Score       string  `json:"score,omitempty"`
	Link        string  `json:"link,omitempty"`
	Image       string  `json:"image,omitempty"`
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Categories  string  `json:"categories,omitempty"`
	Price       float64 `json:"price,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Stars       float64 `json:"starts,omitempty"`

	// metadata
	ScrapUrl  string `json:"scrapUrl,omitempty"`
	ScrapTags string `json:"scrapTags,omitempty"`
	Version   int    `json:"version,omitempty"`
	Index     string `json:"index,omitempty"`
	LastScrap string `json:"lastScrap,omitempty"`
}

/*
// MultiResult represents...
// type MultiResult map[string][]Result
// Entry represents...
type Entry struct {
	id    int `json:"id,omitempty" yaml:"id,omitempty" toml:"id,omitempty"`
	title string `json:"title,omitempty" yaml:"title,omitempty" toml:"title,omitempty"`
	url   string `json:"url,omitempty" yaml:"url,omitempty" toml:"url,omitempty"`
	desc  string `json:"desc,omitempty" yaml:"desc,omitempty" toml:"desc,omitempty"`
}*/
