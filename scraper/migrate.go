package scraper

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/jinzhu/gorm"
	// "github.com/k0kubun/pp"
)

var (
	slugifier = slugify.New(slugify.Configuration{
		ReplaceCharacter: '_',
	})
)

func MigrateTables(db *gorm.DB, isTruncate bool, tables ...interface{}) {
	for _, table := range tables {
		// fmt.Println("table name: ", table)
		if isTruncate {
			if err := db.DropTableIfExists(table).Error; err != nil {
				fmt.Println("table creation error, error msg: ", err)
			}
		}
		//fmt.Println(" ------- START ------- ")
		//pp.Print(table)
		db.AutoMigrate(table)
		//fmt.Println(" ------- END ------- ")
	}
}

// FindOrCreateTagByName finds a tag by name, creating if it doesn't exist
func FindOrCreateProviderByName(db *gorm.DB, name string) (Provider, bool, error) {
	if name == "" {
		return Provider{}, false, errors.New("WARNING !!! No provider name provided")
	}
	var provider Provider
	if db.Where("lower(name) = ?", strings.ToLower(name)).First(&provider).RecordNotFound() {
		provider.Name = name
		err := db.Create(&provider).Error
		return provider, true, err
	}
	return provider, false, nil
}

func FindOrCreateGroupByName(db *gorm.DB, name string) (*Group, bool, error) {
	if name == "" {
		return nil, false, errors.New("WARNING !!! No provider name provided")
	}
	var group Group
	if db.Where("lower(name) = ?", strings.ToLower(name)).First(&group).RecordNotFound() {
		group.Name = name
		err := db.Create(&group).Error
		return &group, true, err
	}
	return &group, false, nil
}

func MigrateEndpoints(db *gorm.DB, c Config) error {
	for _, e := range c.Routes {
		// provider := convertProviderConfig(e.ProviderStr, c.Debug)
		//}
		/*
			if ok := db.NewRecord(provider); ok {
				if err := db.Create(&provider).Error; err != nil {
					fmt.Println("error: ", err)
					return err
				}
			}
		*/

		selectionBlocks, err := convertSelectorsConfig(e.BlocksJSON, c.Debug)
		if err != nil {
			return err
		}
		headers, err := convertHeadersConfig(e.HeadersJSON, c.Debug)
		if err != nil {
			return err
		}

		//if c.Debug {
		//	pp.Print(selectionBlocks)
		//}

		endpointTemplateURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(e.BaseURL, "/"), strings.TrimPrefix(e.PatternURL, "/"))
		slugURL := slugifier.Slugify(endpointTemplateURL)
		// exampleURL := strings.Replace(endpointTemplateURL, "{{query}}", "test", -1)

		endpoint := Endpoint{
			Disabled:   false,
			Route:      e.Route,
			Method:     strings.ToUpper(e.Method),
			BaseURL:    e.BaseURL,
			PatternURL: e.PatternURL,
			Selector:   e.Selector,
			Slug:       slugURL,
			Headers:    headers,
			Blocks:     selectionBlocks,
			// LeafPaths:    e.LeafPaths,
			ExtractPaths: e.ExtractPaths,
			Debug:        e.Debug,
			StrictMode:   e.StrictMode,
		}

		//if c.Debug {
		//	fmt.Printf("\n\nMigrating endpoint: %s/%s \n", e.BaseURL, e.PatternURL)
		// pp.Print(endpoint)
		//}

		var groups []*Group
		group, _, err := FindOrCreateGroupByName(db, "Web")
		if err != nil {
			fmt.Println("Could not upsert the group for the current endpoint. error: ", err)
		}
		groups = append(groups, group)
		// pp.Println("Groups: ", group)
		// pp.Println("Group: ", groups)
		endpoint.Groups = groups

		providerDataURL, err := url.Parse(e.BaseURL)
		if err != nil {
			fmt.Println("Could not parse/extract the endpoint url parts. error: ", err)
			//return err
		}
		providerHost, providerPort, err := net.SplitHostPort(providerDataURL.Host)
		if err != nil {
			// fmt.Println("Could not split host and port for the current endpoint base url. error: ", err)
			// return err
		}

		// if c.Debug {
		// pp.Println(providerDataURL)
		//pp.Println(providerHost)
		//pp.Println(providerPort)
		// }
		providerDomain := domainutil.Domain(providerDataURL.Host)
		//if c.Debug {
		//	pp.Println(providerDomain)
		//}

		if providerHost != "" {
			endpoint.Host = providerHost
		} else {
			endpoint.Host = providerDataURL.Host
		}
		endpoint.Domain = providerDomain

		if providerPort != "" {
			providerPortInt, err := strconv.Atoi(providerPort)
			if err != nil {
				fmt.Println("WARNING ! Missing the port number for this endpoint base url. error: ", err)
			}
			endpoint.Port = providerPortInt
		} else {
			// Move to a seperate method
			switch providerDataURL.Scheme {
			case "wss":
			case "https":
				endpoint.Port = 443
			case "ws":
			case "http":
				endpoint.Port = 80
			case "rpc":
				endpoint.Port = 445
			default:
				fmt.Println("WARNING ! invalid base url scheme for the current endpoint.")
			}
		}

		endpoint.Description = e.Description

		//pp.Print(endpoint)
		provider, _, err := FindOrCreateProviderByName(db, providerDomain)
		if err != nil {
			fmt.Println("Could not upsert the current provider in the registry. error: ", err)
			// return err
		}

		// endpoint.ProviderID = provider.ID
		// pp.Print(provider)
		endpoint.Provider = provider

		for _, b := range selectionBlocks {
			if ok := db.NewRecord(b); ok {
				if err := db.Create(&b).Error; err != nil {
					fmt.Println("error: ", err)
					return err
				}
			}
		}

		if ok := db.NewRecord(endpoint); ok {
			if err := db.Create(&endpoint).Error; err != nil {
				fmt.Println("error: ", err)
				return err
			}
		}

	}
	return nil
}

func convertProviderConfig(name string, debug bool) *Provider {
	provider := &Provider{}
	if name != "" {
		provider.Name = name
	} else {
		return nil
	}
	if debug {
		fmt.Printf("\nConverting provider name: '%s' \n", name)
	}
	return provider
}

func convertSelectorsConfig(selectors map[string]SelectorConfig, debug bool) ([]*SelectorConfig, error) {
	var blocks []*SelectorConfig
	for k, v := range selectors {
		targets, err := convertDetailsConfig(v.Details, debug)
		if err != nil {
			return nil, err
		}
		selection := &SelectorConfig{
			Collection: k,
			Debug:      v.Debug,
			Required:   v.Required,
			Items:      v.Items,
			Matchers:   targets,
			StrictMode: v.StrictMode,
		}
		//if debug {
		//	fmt.Printf("\nConverting selector config: %s \n", k)
		//fmt.Println("Input:")
		//pp.Print(v)
		//fmt.Println("Output:")
		//pp.Print(selection)
		//}
		blocks = append(blocks, selection)
	}
	return blocks, nil
}

func convertDetailsConfig(tgts map[string]Extractors, debug bool) ([]*MatcherConfig, error) {
	var targets []*MatcherConfig
	for k, t := range tgts {
		var matchers []Matcher
		for _, e := range t {
			matchers = append(matchers, Matcher{Expression: e.val})
		}
		target := &MatcherConfig{
			Target:  k,
			Selects: matchers,
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func convertHeadersConfig(headers map[string]string, debug bool) ([]*HeaderConfig, error) {
	var hdrs []*HeaderConfig
	for k, v := range headers {
		header := &HeaderConfig{
			Key:   k,
			Value: v,
		}
		if debug {
			fmt.Printf("\nConverting header config: %s:%s \n", k, v)
		}
		hdrs = append(hdrs, header)
	}
	return hdrs, nil
}

func createGroups(db *gorm.DB) {
	for _, g := range Seeds.Groups {
		group := Group{}
		group.Name = g.Name
		if err := db.Create(&group).Error; err != nil {
			log.Fatalf("create group (%v) failure, got err %v", group, err)
		}
	}
}

func createTopics(db *gorm.DB) {
	for _, t := range Seeds.Topics {
		topic := Topic{}
		topic.Name = t.Name
		topic.Code = strings.ToLower(t.Name)
		if err := db.Create(&topic).Error; err != nil {
			log.Fatalf("create topic (%v) failure, got err %v", topic, err)
		}
	}
}
