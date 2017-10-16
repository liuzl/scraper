package scraper

import (
	"fmt"
	"log"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
)

func MigrateTables(db *gorm.DB, isTruncate bool, tables ...interface{}) {
	for _, table := range tables {
		if isTruncate {
			if err := db.DropTableIfExists(table).Error; err != nil {
				panic(err)
			}
		}
		db.AutoMigrate(table)
	}
}

func MigrateEndpoints(db *gorm.DB, c Config) error {
	for _, e := range c.Routes {
		provider := convertProviderConfig(e.ProviderStr, c.Debug)
		selectionBlocks, err := convertSelectorsConfig(e.Blocks, c.Debug)
		if err != nil {
			return err
		}
		headers, err := convertHeadersConfig(e.Headers, c.Debug)
		if err != nil {
			return err
		}
		pp.Print(selectionBlocks)
		endpoint := Endpoint{
			Disabled: false,
			// Provider:   provider,
			Route:      e.Route,
			Name:       e.Name,
			Method:     e.Method,
			BaseURL:    e.BaseURL,
			PatternURL: e.PatternURL,
			Body:       e.Body,
			Selector:   e.Selector,
			HeadersORM: headers,
			BlocksORM:  selectionBlocks,
			// Extract:    ExtractConfig{},
			Debug:      e.Debug,
			StrictMode: e.StrictMode,
		}
		/*
			if c.Debug {
				fmt.Printf("\n\nMigrating endpoint: %s \n", e.Name)
				pp.Print(endpoint)
			}
		*/
		if ok := db.NewRecord(provider); ok {
			if err := db.Create(&provider).Error; err != nil {
				fmt.Println("error: ", err)
				return err
			}
		}

		for _, b := range selectionBlocks {
			//d := b.DetailsORM
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

		//endpoint.HeadersORM = headers
		//endpoint.BlocksORM = selectionBlocks

		//if ok := db.NewRecord(endpoint); ok {
		//if err := db.Create(&endpoint).Error; err != nil {
		//	fmt.Println("error: ", err)
		//	return err
		//}
		//}

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
	/*
		if debug {
			fmt.Printf("\nConverting provider name: '%s' \n", name)
		}
	*/
	return provider
}

func convertSelectorsConfig(selectors map[string]SelectorConfig, debug bool) ([]SelectorConfig, error) {
	var blocks []SelectorConfig
	for _, v := range selectors {
		targets, err := convertDetailsConfig(v.Details, debug)
		if err != nil {
			return nil, err
		}
		selection := SelectorConfig{
			Slug:       v.Slug,
			Debug:      v.Debug,
			Required:   v.Required,
			Selector:   v.Selector,
			Items:      v.Items,
			DetailsORM: targets,
			StrictMode: v.StrictMode,
		}
		/*
			if debug {
				fmt.Printf("\nConverting selector config: %s \n", k)
				fmt.Println("Input:")
				pp.Print(v)
				fmt.Println("Output:")
				pp.Print(selection)
			}
		*/
		blocks = append(blocks, selection)
	}
	return blocks, nil
}

func convertDetailsConfig(tgts map[string]Extractors, debug bool) ([]ExtractorORM, error) {
	var targets []ExtractorORM
	for k, t := range tgts {
		// []*Extractor
		for c, e := range t {
			target := ExtractorORM{
				Target:  k,
				Matcher: e.val,
			}
			if debug {
				fmt.Printf("\nConverting extractor target config: '%s' = '%s' \n", k, e.val)
				pp.Print(c)
				pp.Print(t)
			}
			targets = append(targets, target)
		}
	}
	return targets, nil
}

func convertHeadersConfig(headers map[string]string, debug bool) ([]HeaderConfig, error) {
	var hdrs []HeaderConfig
	for k, v := range headers {
		header := HeaderConfig{
			Key:   k,
			Value: v,
		}
		/*
			if debug {
				fmt.Printf("\nConverting header config: %s:%s \n", k, v)
			}
		*/
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
