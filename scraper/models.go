package scraper

// Result represents a result
type Result map[string]string

// Config represents...
type Config struct {
	Port   int         `default:"3000" json:"port,omitempty" yaml:"port,omitempty" toml:"port,omitempty"`
	Routes []*Endpoint `json:"routes,omitempty" yaml:"routes,omitempty" toml:"routes,omitempty"`
}

// Endpoint represents a single remote endpoint. The performed query can be modified between each call by parameterising URL. See documentation.
type Endpoint struct {
	Disabled   bool `default:"false" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Debug      bool `default:"false" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	StrictMode bool `default:"false" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`

	Route   string `json:"route,omitempty" yaml:"route,omitempty" toml:"route,omitempty"`
	Name    string `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	Method  string `json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`
	BaseURL string `json:"base_url,omitempty" yaml:"base_url,omitempty" toml:"base_url,omitempty"`
	URL     string `json:"url" yaml:"url" toml:"url"`
	Body    string `json:"body,omitempty" yaml:"body,omitempty" toml:"body,omitempty"`

	Selector  string                    `default:"css" json:"selector,omitempty" yaml:"selector,omitempty" toml:"selector,omitempty"`
	Headers   map[string]string         `json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
	Blocks    map[string]SelectorConfig `json:"blocks,omitempty" yaml:"blocks,omitempty" toml:"blocks,omitempty"`
	Extract   ExtractConfig             `default:"false" json:"extract,omitempty" yaml:"extract,omitempty" toml:"extract,omitempty"`
	MinFields int                       `json:"-" yaml:"-" toml:"-"`
	Count     string                    `json:"-" yaml:"-" toml:"-"`
}

// SelectorConfig represents a content selection rule for a single URL Pattern.
type SelectorConfig struct {
	Slug       string                `json:"slug,omitempty" yaml:"slug,omitempty" toml:"slug,omitempty"`
	Debug      bool                  `default:"true" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	Required   bool                  `default:"true" json:"required,omitempty" yaml:"required,omitempty" toml:"required,omitempty"`
	Selector   string                `default:"css" json:"selector,omitempty" yaml:"selector,omitempty" toml:"selector,omitempty"`
	Items      string                `json:"items,omitempty" yaml:"items,omitempty" toml:"items,omitempty"`
	Details    map[string]Extractors `json:"details,omitempty" yaml:"details,omitempty" toml:"details,omitempty"`
	StrictMode bool                  `default:"false" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
}

// Extractor represents a pair of css selector and extracted node.
type Extractor struct {
	val string
	fn  extractorFn
}

// ExtractConfig represents a single sub-extraction rules url content configuration.
type ExtractConfig struct {
	Debug     bool `default:"true" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	Links     bool `default:"true" json:"links,omitempty" yaml:"links,omitempty" toml:"links,omitempty"`
	Meta      bool `default:"true" json:"meta,omitempty" yaml:"meta,omitempty" toml:"meta,omitempty"`
	OpenGraph bool `default:"true" json:"opengraph,omitempty" yaml:"opengraph,omitempty" toml:"opengraph,omitempty"`
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
