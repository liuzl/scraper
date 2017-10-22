package scraper

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/media_library"
	"github.com/qor/sorting"
	"github.com/qor/validations"
)

// WEB SCRAPER ///////////////////////////////////////////////////////////////

// Result represents a result
type Result map[string]interface{}

// Create a GORM-backend model
type Provider struct {
	gorm.Model      `json:"-" yaml:"-" toml:"-"`
	sorting.Sorting `json:"-" yaml:"-" toml:"-"`
	// ProviderID uint
	Name  string                   `etcd:"name" required:"true" json:"name" yaml:"name" toml:"name"` // gorm:"type:varchar(128);unique_index"
	Logo  media_library.MediaBox   `json:"-" yaml:"-" toml:"-"`
	Ranks []*ProviderWebRankConfig `json:"ranks,omitempty" yaml:"ranks,omitempty" toml:"ranks,omitempty"`
	// Endpoints []*Endpoint              `json:"endpoints,omitempty" yaml:"endpoints,omitempty" toml:"endpoints,omitempty"`
}

type ProviderWebRankConfig struct {
	gorm.Model      `json:"-" yaml:"-" toml:"-"`
	sorting.Sorting `json:"-" yaml:"-" toml:"-"`
	ProviderID      uint   `json:"-" yaml:"-" toml:"-"`
	Engine          string `json:"engine,omitempty" yaml:"engine,omitempty" toml:"engine,omitempty"`
	Score           string `json:"score,omitempty" yaml:"score,omitempty" toml:"score,omitempty"`
}

// Config represents...
type Config struct {
	gorm.Model      `json:"-" yaml:"-" toml:"-"`
	sorting.Sorting `json:"-" yaml:"-" toml:"-"`
	Disabled        bool        `default:"false" help:"Disable handler init" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Env             EnvConfig   `gorm:"-" json:"env,omitempty" yaml:"env,omitempty" toml:"env,omitempty"`
	Etcd            EtcdConfig  `opts:"-" json:"etcd,omitempty" yaml:"etcd,omitempty" toml:"etcd,omitempty"`
	Port            int         `default:"3000" json:"port,omitempty" yaml:"port,omitempty" toml:"port,omitempty"`
	Dashboard       bool        `default:"false" help:"Initialize the Administration Interface" json:"dashboard,omitempty" yaml:"dashboard,omitempty" toml:"dashboard,omitempty"`
	Truncate        bool        `default:"true" help:"Truncate previous data" json:"truncate,omitempty" yaml:"truncate,omitempty" toml:"truncate,omitempty"`
	Migrate         bool        `default:"true" help:"Migrate to admin dashboard" json:"migrate,omitempty" yaml:"migrate,omitempty" toml:"migrate,omitempty"`
	Debug           bool        `default:"false" help:"Enable debug output" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	Routes          []*Endpoint `gorm:"-" json:"routes,omitempty" yaml:"routes,omitempty" toml:"routes,omitempty"`
}

type EnvConfig struct {
	Disabled      bool                         `default:"false" help:"Disable handler init" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Files         []string                     `json:"files,omitempty" yaml:"files,omitempty" toml:"files,omitempty"`
	VariablesList map[string]string            `json:"-" yaml:"-" toml:"-"`
	VariablesTree map[string]map[string]string `json:"-" yaml:"-" toml:"-"`
	Debug         bool                         `default:"false" help:"Enable debug output for env vars processing" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

// Endpoint represents a single remote endpoint. The performed query can be modified between each call by parameterising URL. See documentation.
type Endpoint struct {
	gorm.Model         `json:"-" yaml:"-" toml:"-"`
	sorting.Sorting    `json:"-" yaml:"-" toml:"-"`
	Update             time.Time                    `json:"-" yaml:"-" toml:"-"`
	Disabled           bool                         `etcd:"disabled" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Loaded             bool                         `json:"-" yaml:"-" toml:"-"`
	EtcdKey            string                       `etcd:"etcd_key" json:"etcd_key,omitempty" yaml:"etcd_key,omitempty" toml:"etcd_key,omitempty"`
	Connections        []Connection                 `json:"-" yaml:"-" toml:"-"`
	Source             string                       `etcd:"source" gorm:"-" json:"provider,omitempty" yaml:"provider,omitempty" toml:"provider,omitempty"`
	ProviderID         uint                         `json:"-" yaml:"-" toml:"-"`
	Provider           Provider                     `etcd:"provider" json:"provider_orm,omitempty" yaml:"provider_orm,omitempty" toml:"provider_orm,omitempty"`
	Comment            string                       `json:"comments,omitempty" yaml:"comments,omitempty" toml:"comments,omitempty"`
	Description        string                       `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Groups             []*Group                     `etcd:"groups" json:"groups,omitempty" yaml:"groups,omitempty" toml:"groups,omitempty"`
	Route              string                       `etcd:"router" json:"route,omitempty" yaml:"route,omitempty" toml:"route,omitempty"`
	Method             string                       `gorm:"index" json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`
	Domain             string                       `gorm:"-" json:"-" yaml:"-" toml:"-"`
	Host               string                       `gorm:"-" json:"-" yaml:"-" toml:"-"`
	Port               int                          `gorm:"-" json:"-" yaml:"-" toml:"-"`
	BaseURL            string                       `etcd:"base_url" gorm:"index" json:"base_url,omitempty" yaml:"base_url,omitempty" toml:"base_url,omitempty"`
	PatternURL         string                       `etcd:"url" json:"url" yaml:"url" toml:"url"`
	Examples           map[string]map[string]string `gorm:"-" json:"examples" yaml:"examples" toml:"examples"`
	Slug               string                       `etcd:"slug" json:"slug,omitempty" yaml:"slug,omitempty" toml:"slug,omitempty"`
	ExtractPaths       bool                         `etcd:"extract_paths" json:"extract_paths,omitempty" yaml:"extract_paths,omitempty" toml:"extract_paths,omitempty"`
	LeafPaths          []string                     `gorm:"-" json:"leaf_paths,omitempty" yaml:"leaf_paths,omitempty" toml:"leaf_paths,omitempty"`
	Body               string                       `gorm:"-" json:"body,omitempty" yaml:"body,omitempty" toml:"body,omitempty"`
	Selector           string                       `etcd:"selector" gorm:"index" default:"css" json:"selector,omitempty" yaml:"selector,omitempty" toml:"selector,omitempty"`
	HeadersIntercept   []string                     `etcd:"resp_headers_intercept" gorm:"-" json:"resp_headers_intercept,omitempty" yaml:"resp_headers_intercept,omitempty" toml:"resp_headers_intercept,omitempty"`
	HeadersJSON        map[string]string            `etcd:"headers" gorm:"-" json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
	BlocksJSON         map[string]SelectorConfig    `etcd:"blocks" gorm:"-" json:"blocks,omitempty" yaml:"blocks,omitempty" toml:"blocks,omitempty"`
	Headers            []*HeaderConfig              `json:"headers_orm,omitempty" yaml:"headers_orm,omitempty" toml:"headers_orm,omitempty"`
	Blocks             []*SelectorConfig            `json:"blocks_orm,omitempty" yaml:"blocks_orm,omitempty" toml:"blocks_orm,omitempty"`
	EndpointProperties EndpointProperties           `etcd:"properties" sql:"type:text" json:"properties,omitempty" yaml:"properties,omitempty" toml:"properties,omitempty"`
	Extract            ExtractConfig                `etcd:"extract" default:"false" json:"extract,omitempty" yaml:"extract,omitempty" toml:"extract,omitempty"`
	MinFields          int                          `json:"-" yaml:"-" toml:"-"`
	Count              string                       `gorm"-" json:"-" yaml:"-" toml:"-"`
	Debug              bool                         `etcd:"debug" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	StrictMode         bool                         `etcd:"strict_mode" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
	// Screenshot  Screenshot `json:"-" yaml:"-" toml:"-"`
}

type Screenshot struct {
	gorm.Model   `json:"-" yaml:"-" toml:"-"`
	Title        string                            `etcd:"title" json:"title,omitempty" yaml:"title,omitempty" toml:"title,omitempty"`
	EndpointID   uint                              `json:"-" yaml:"-" toml:"-"`
	SelectedType string                            `etcd:"selected_type" json:"selected_type,omitempty" yaml:"selected_type,omitempty" toml:"selected_type,omitempty"`
	File         media_library.MediaLibraryStorage `sql:"size:4294967295;" media_library:"url:/system/{{class}}/{{primary_key}}/{{column}}.{{extension}}" json:"-" yaml:"-" toml:"-"`
	// Category     Category
	// CategoryID   uint
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
	gorm.Model          `json:"-" yaml:"-" toml:"-"`
	sorting.SortingDESC `json:"-" yaml:"-" toml:"-"`
	Keywords            []Query `etcd:"keywords" json:"keywords,omitempty" yaml:"keywords,omitempty" toml:"keywords,omitempty"`
}

type Query struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	InputQuery string `etcd:"input_query" json:"input_query,omitempty" yaml:"input_query,omitempty" toml:"input_query,omitempty"`
	Slug       string `etcd:"slug" json:"slug,omitempty" yaml:"slug,omitempty" toml:"slug,omitempty"`
	MD5        string `etcd:"md5" json:"md5,omitempty" yaml:"md5,omitempty" toml:"md5,omitempty"`
	SHA1       string `etcd:"sha1" json:"sha1,omitempty" yaml:"sha1,omitempty" toml:"sha1,omitempty"`
	UUID       string `etcd:"uuid" json:"uuid,omitempty" yaml:"uuid,omitempty" toml:"uuid,omitempty"`
	Blocked    bool   `etcd:"blocked" json:"blocked,omitempty" yaml:"blocked,omitempty" toml:"blocked,omitempty"`
}

type EndpointProperties []EndpointProperty // `etcd:"properties" json:"properties" yaml:"properties" toml:"properties"`

type EndpointProperty struct {
	Name  string `etcd:"name" json:"name" yaml:"name" toml:"name"`
	Value string `etcd:"value" json:"value" yaml:"value" toml:"value"`
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
	gorm.Model      `json:"-" yaml:"-" toml:"-"`
	sorting.Sorting `json:"-" yaml:"-" toml:"-"`
	EndpointID      uint                  `json:"-" yaml:"-" toml:"-"`
	EtcdKey         string                `etcd:"etcd_key" json:"etcd_key,omitempty" yaml:"etcd_key,omitempty" toml:"etcd_key,omitempty"`
	Collection      string                `json:"collection,omitempty" yaml:"collection,omitempty" toml:"collection,omitempty"`
	Description     string                `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Required        bool                  `etcd:"required" default:"true" json:"required,omitempty" yaml:"required,omitempty" toml:"required,omitempty"`
	Items           string                `etcd:"items" json:"items,omitempty" yaml:"items,omitempty" toml:"items,omitempty"`
	Details         map[string]Extractors `etcd:"details" gorm:"-" json:"details,omitempty" yaml:"details,omitempty" toml:"details,omitempty"`
	Paths           map[string]string     `etcd:"paths" gorm:"-" json:"paths,omitempty" yaml:"paths,omitempty" toml:"paths,omitempty"`
	Matchers        []*MatcherConfig      `json:"matchers,omitempty" yaml:"matchers,omitempty" toml:"matchers,omitempty"`
	StrictMode      bool                  `etcd:"strict_mode" default:"false" json:"strict_mode,omitempty" yaml:"strict_mode,omitempty" toml:"strict_mode,omitempty"`
	Debug           bool                  `etcd:"debug" default:"true" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

// Extractor represents a pair of css selector and extracted node.
type Extractor struct {
	val string      `etcd:"value" json:"value" yaml:"value" toml:"value"`
	fn  extractorFn `gorm:"-" json:"-" yaml:"-" toml:"-"`
}

// Extractor represents a pair of css selector and extracted node.
type MatcherConfig struct {
	gorm.Model       `json:"-" yaml:"-" toml:"-"`
	sorting.Sorting  `json:"-" yaml:"-" toml:"-"`
	SelectorConfigID uint      `json:"-" yaml:"-" toml:"-"`
	Target           string    `etcd:"target" json:"target,omitempty" yaml:"target,omitempty" toml:"target,omitempty"`
	Selects          []Matcher `etcd:"selects" json:"selects,omitempty" yaml:"selects,omitempty" toml:"selects,omitempty"`
	EtcdKey          string    `etcd:"etcd_key" json:"etcd_key,omitempty" yaml:"etcd_key,omitempty" toml:"etcd_key,omitempty"`
}

//type Matchers {[]Matcher
type Matcher struct {
	gorm.Model      `json:"-" yaml:"-" toml:"-"`
	MatcherConfigID uint   `json:"-" yaml:"-" toml:"-"`
	Expression      string `etcd:"expr" json:"expr,omitempty" yaml:"expr,omitempty" toml:"expr,omitempty"`
}

type SelectorType struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	Name       string `etcd:"name" json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	Engine     string `etcd:"engine" json:"engine,omitempty" yaml:"engine,omitempty" toml:"engine,omitempty"`
}

type TargetConfig struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	// EndpointID uint `json:"-" yaml:"-" toml:"-"`
	Name string `etcd:"name" json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
}

type HeaderConfig struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	EndpointID uint   `json:"-" yaml:"-" toml:"-"`
	Key        string `etcd:"key" json:"key,omitempty" yaml:"key,omitempty" toml:"key,omitempty"`
	Value      string `etcd:"value" json:"value,omitempty" yaml:"value,omitempty" toml:"value,omitempty"`
}

type BlocksConfig struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	Key        string         `etcd:"key" json:"key,omitempty" yaml:"key,omitempty" toml:"key,omitempty"`
	Value      SelectorConfig `etcd:"value" json:"value,omitempty" yaml:"value,omitempty" toml:"value,omitempty"`
}

type ExtractorsConfig struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	Key        string     `etcd:"key" json:"key,omitempty" yaml:"key,omitempty" toml:"key,omitempty"`
	Value      Extractors `etcd:"value" json:"value,omitempty" yaml:"value,omitempty" toml:"value,omitempty"`
}

// ExtractConfig represents a single sub-extraction rules url content configuration.
type ExtractConfig struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	Debug      bool `default:"true" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
	Links      bool `default:"true" json:"links,omitempty" yaml:"links,omitempty" toml:"links,omitempty"`
	Meta       bool `default:"true" json:"meta,omitempty" yaml:"meta,omitempty" toml:"meta,omitempty"`
	Opengraph  bool `default:"true" json:"opengraph,omitempty" yaml:"opengraph,omitempty" toml:"opengraph,omitempty"`
}

// OPENAPI SCRAPER ///////////////////////////////////////////////////////////////
type OpenAPIConfig struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	Name       string                `etcd:"name" json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	Provider   Provider              `etcd:"provider" json:"provider,omitempty" yaml:"provider,omitempty" toml:"provider,omitempty"`
	Specs      []*OpenAPISpecsConfig `etcd:"specs" json:"specs,omitempty" yaml:"specs,omitempty" toml:"specs,omitempty"`
}

type OpenAPISpecsConfig struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	Slug       string `etcd:"slug" json:"slug,omitempty" yaml:"slug,omitempty" toml:"slug,omitempty"`
	Version    string `etcd:"version" json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`
}

// REQUESTS API WEBMOCKS ///////////////////////////////////////////////////////////////
type Connection struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	// ID         uint     `gorm:"primary_key;AUTO_INCREMENT" json:"-" yaml:"-" toml:"-"`
	EndpointID uint     `json:"-" yaml:"-" toml:"-"`
	URL        string   `json:"url" yaml:"url" toml:"url"`
	Request    Request  `json:"request" yaml:"request" toml:"request"`
	Response   Response `json:"response" yaml:"response" toml:"response"`
	Provider   Provider `json:"provider" yaml:"provider" toml:"provider"`
	RecordedAt string   `json:"recorded_at" yaml:"recorded_at" toml:"recorded_at"`
}

type Request struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	// ID           uint   `gorm:"primary_key;AUTO_INCREMENT" json:"-" yaml:"-" toml:"-"`
	ConnectionID uint   `json:"-" yaml:"-" toml:"-"`
	Header       string `json:"header" yaml:"header" toml:"header"`
	Body         string `json:"body" yaml:"body" toml:"body"`
	Method       string `json:"method" yaml:"method" toml:"method"`
	URL          string `json:"url" yaml:"url" toml:"url"`
}

type Response struct {
	gorm.Model `json:"-" yaml:"-" toml:"-"`
	// ID           uint   `gorm:"primary_key;AUTO_INCREMENT" json:"-" yaml:"-" toml:"-"`
	ConnectionID uint   `json:"-" yaml:"-" toml:"-"`
	Status       string `json:"status" yaml:"status" toml:"status"`
	Header       string `json:"header" yaml:"header" toml:"header"`
	Body         string `json:"body" yaml:"body" toml:"body"`
}
