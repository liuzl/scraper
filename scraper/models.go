package scraper

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/media_library"
	"github.com/qor/sorting"
	"github.com/qor/validations"
	// "github.com/qor/publish2"
	// "github.com/qor/media"
	// "github.com/qor/media/oss"
)

// WEB SCRAPER ///////////////////////////////////////////////////////////////

// Result represents a result
type Result map[string]string

// Create a GORM-backend model
type Provider struct {
	gorm.Model
	sorting.Sorting
	// ProviderID uint
	Name  string                   `required:"true" json:"name" yaml:"name" toml:"name"` // gorm:"type:varchar(128);unique_index"
	Logo  media_library.MediaBox   `json:"-" yaml:"-" toml:"-"`
	Ranks []*ProviderWebRankConfig `json:"ranks,omitempty" yaml:"ranks,omitempty" toml:"ranks,omitempty"`
	// Endpoints []*Endpoint              `json:"endpoints,omitempty" yaml:"endpoints,omitempty" toml:"endpoints,omitempty"`
}

type ProviderWebRankConfig struct {
	gorm.Model
	sorting.Sorting
	ProviderID uint
	Engine     string `json:"google,omitempty" yaml:"google,omitempty" toml:"google,omitempty"`
	Score      string `json:"bing,omitempty" yaml:"bing,omitempty" toml:"bing,omitempty"`
}

// Config represents...
type Config struct {
	gorm.Model
	sorting.Sorting
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
	sorting.Sorting
	Disabled bool `default:"false" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`

	Source      string   `gorm:"-" json:"provider,omitempty" yaml:"provider,omitempty" toml:"provider,omitempty"`
	ProviderID  uint     `json:"-" yaml:"-" toml:"-"`
	Provider    Provider `json:"provider_orm,omitempty" yaml:"provider_orm,omitempty" toml:"provider_orm,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Groups      []*Group `json:"groups,omitempty" yaml:"groups,omitempty" toml:"groups,omitempty"`
	// Screenshot  Screenshot `json:"-" yaml:"-" toml:"-"` // media_library.MediaBox `json:"-" yaml:"-" toml:"-"`

	Route  string `json:"route,omitempty" yaml:"route,omitempty" toml:"route,omitempty"`
	Method string `gorm:"index" json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`

	Domain string `gorm:"-" json:"-" yaml:"-" toml:"-"`
	Host   string `gorm:"-" json:"-" yaml:"-" toml:"-"`
	Port   int    `gorm:"-" json:"-" yaml:"-" toml:"-"`

	BaseURL    string `gorm:"index" json:"base_url,omitempty" yaml:"base_url,omitempty" toml:"base_url,omitempty"`
	PatternURL string `json:"url" yaml:"url" toml:"url"`
	ExampleURL string `json:"example_url" yaml:"example_url" toml:"example_url"`
	Slug       string `json:"slug,omitempty" yaml:"slug,omitempty" toml:"slug,omitempty"`

	Body     string `gorm:"-" json:"body,omitempty" yaml:"body,omitempty" toml:"body,omitempty"`
	Selector string `gorm:"index" default:"css" json:"selector,omitempty" yaml:"selector,omitempty" toml:"selector,omitempty"`

	HeadersJSON map[string]string         `gorm:"-" json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
	BlocksJSON  map[string]SelectorConfig `gorm:"-" json:"blocks,omitempty" yaml:"blocks,omitempty" toml:"blocks,omitempty"`

	Headers []*HeaderConfig   `json:"headers_orm,omitempty" yaml:"headers_orm,omitempty" toml:"headers_orm,omitempty"`
	Blocks  []*SelectorConfig `json:"blocks_orm,omitempty" yaml:"blocks_orm,omitempty" toml:"blocks_orm,omitempty"`

	EndpointProperties EndpointProperties `sql:"type:text" json:"properties,omitempty" yaml:"properties,omitempty" toml:"properties,omitempty"`

	Extract   ExtractConfig `default:"false" json:"extract,omitempty" yaml:"extract,omitempty" toml:"extract,omitempty"`
	MinFields int           `json:"-" yaml:"-" toml:"-"`
	Count     string        `gorm"-" json:"-" yaml:"-" toml:"-"`

	Debug      bool `default:"false" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	StrictMode bool `default:"false" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
}

type Screenshot struct {
	gorm.Model
	Title      string
	EndpointID uint `json:"-" yaml:"-" toml:"-"`
	// Category     Category
	// CategoryID   uint
	SelectedType string
	File         media_library.MediaLibraryStorage `sql:"size:4294967295;" media_library:"url:/system/{{class}}/{{primary_key}}/{{column}}.{{extension}}"`
}

func (screenshot Screenshot) Validate(db *gorm.DB) {
	if strings.TrimSpace(screenshot.Title) == "" {
		db.AddError(validations.NewError(screenshot, "Title", "Title can not be empty"))
	}
}

func (screenshot *Screenshot) SetSelectedType(typ string) {
	screenshot.SelectedType = typ
}

func (screenshot *Screenshot) GetSelectedType() string {
	return screenshot.SelectedType
}

func (screenshot *Screenshot) ScanMediaOptions(mediaOption media_library.MediaOption) error {
	if bytes, err := json.Marshal(mediaOption); err == nil {
		return screenshot.File.Scan(bytes)
	} else {
		return err
	}
}

func (screenshot *Screenshot) GetMediaOption() (mediaOption media_library.MediaOption) {
	mediaOption.Video = screenshot.File.Video
	mediaOption.FileName = screenshot.File.FileName
	mediaOption.URL = screenshot.File.URL()
	mediaOption.OriginalURL = screenshot.File.URL("original")
	mediaOption.CropOptions = screenshot.File.CropOptions
	mediaOption.Sizes = screenshot.File.GetSizes()
	mediaOption.Description = screenshot.File.Description
	return
}

/*
type ScreenShotVariationImageStorage struct{ oss.OSS }

func (colorVariation ScreenShot) MainImageURL() string {
	if len(colorVariation.Images.Files) > 0 {
		return colorVariation.Images.URL()
	}
	return "/images/default_product.png"
}

func (ScreenShotVariationImageStorage) GetSizes() map[string]*media.Size {
	return map[string]*media.Size{
		"small":  {Width: 320, Height: 320},
		"middle": {Width: 640, Height: 640},
		"big":    {Width: 1280, Height: 1280},
	}
}
*/

type Queries struct {
	gorm.Model
	sorting.SortingDESC
	Keywords []Query
}

type Query struct {
	gorm.Model
	InputQuery string
	Slug       string
	MD5        string
	SHA1       string
	UUID       string
	Blocked    bool
}

type EndpointProperties []EndpointProperty

type EndpointProperty struct {
	Name  string
	Value string
}

func (endpointProperties *EndpointProperties) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, endpointProperties)
	case string:
		if v != "" {
			return endpointProperties.Scan([]byte(v))
		}
	default:
		return errors.New("not supported")
	}
	return nil
}

func (endpointProperties EndpointProperties) Value() (driver.Value, error) {
	if len(endpointProperties) == 0 {
		return nil, nil
	}
	return json.Marshal(endpointProperties)
}

// SelectorConfig represents a content selection rule for a single URL Pattern.
type SelectorConfig struct {
	gorm.Model
	sorting.Sorting
	EndpointID  uint                  `json:"-" yaml:"-" toml:"-"`
	Collection  string                `json:"collection,omitempty" yaml:"collection,omitempty" toml:"collection,omitempty"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Required    bool                  `default:"true" json:"required,omitempty" yaml:"required,omitempty" toml:"required,omitempty"`
	Items       string                `json:"items,omitempty" yaml:"items,omitempty" toml:"items,omitempty"`
	Details     map[string]Extractors `gorm:"-" json:"details,omitempty" yaml:"details,omitempty" toml:"details,omitempty"`
	Matchers    []*MatcherConfig      `json:"matchers,omitempty" yaml:"matchers,omitempty" toml:"matchers,omitempty"`
	StrictMode  bool                  `default:"false" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
	Debug       bool                  `default:"true" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

// Extractor represents a pair of css selector and extracted node.
type Extractor struct {
	val string
	fn  extractorFn `gorm:"-"`
}

// Extractor represents a pair of css selector and extracted node.
type MatcherConfig struct {
	gorm.Model
	sorting.Sorting
	SelectorConfigID uint
	Target           string // TargetConfig
	Selects          []Matcher
	// fn  extractorFn
}

//type Matchers {[]Matcher
type Matcher struct {
	gorm.Model
	MatcherConfigID uint
	Expression      string
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

// OPENAPI SCRAPER ///////////////////////////////////////////////////////////////
type OpenAPIConfig struct {
	gorm.Model
	Name     string
	Provider Provider
	Specs    []*OpenAPISpecsConfig
}

type OpenAPISpecsConfig struct {
	gorm.Model
	Slug    string
	Version string
}
