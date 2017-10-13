package scraper

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	// "strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/xquery/html"
	"github.com/k0kubun/pp"
	"golang.org/x/net/html"
	// "github.com/antchfx/xpath"
	// "github.com/advancedlogic/GoOse"
	/*
		"github.com/ynqa/word-embedding/builder"
		"github.com/ynqa/word-embedding/config"
		"github.com/ynqa/word-embedding/validate"
	*/)

/*
	Refs:
	- github.com/slotix/dataflowkit
	- github.com/slotix/pageres-go-wrapper
	- github.com/fern4lvarez/go-metainspector
	- github.com/gpahal/go-meta
	- https://github.com/scrapinghub/mdr
	- https://github.com/scrapinghub/aile/blob/master/demo2.py
	- https://github.com/datatogether/sentry
	- https://github.com/sourcegraph/webloop
	- https://github.com/107192468/sp/blob/master/src/readhtml/readhtml.go
	- https://github.com/nikolay-turpitko/structor
	- https://github.com/dreampuf/paw/tree/master/src/web
	- https://github.com/rakanalh/grawler/blob/master/processors/text.go
	- https://github.com/rakanalh/grawler/blob/master/extractor/xpath.go
	- https://github.com/rakanalh/grawler/blob/master/extractor/css.go
	- https://github.com/ErosZy/labour/blob/master/parser/pageItemXpathParser.go
	- https://github.com/ErosZy/labour
	- https://github.com/cugbliwei/crawler/blob/master/extractor/selector.go
	- https://github.com/xlvector/higgs/blob/master/extractor/selector.go
	- github.com/tchssk/link
	- https://github.com/peterhellberg/link
*/

// Endpoint represents a single remote endpoint. The performed query can be modified between each call by parameterising URL. See documentation.
type Endpoint struct {
	Disabled   bool `default:"false" json:"disabled,omitempty"`
	Debug      bool `default:"false" json:"debug,omitempty"`
	StrictMode bool `default:"false" json:"strict_mode,omitempty"`

	Route   string `json:"route,omitempty"`
	Name    string `json:"name,omitempty"`
	Method  string `json:"method,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
	URL     string `json:"url"`
	Body    string `json:"body,omitempty"`

	Selector  string                    `default:"css" json:"selector,omitempty"`
	Headers   map[string]string         `json:"headers,omitempty"`
	Blocks    map[string]SelectorConfig `json:"blocks,omitempty"`
	Extract   ExtractConfig             `default:"false" json:"extract,omitempty"`
	MinFields int                       `json:"-"`
	Count     string                    `json:"-"`
}

type ExtractConfig struct {
	Debug     bool `default:"true" json:"debug,omitempty"`
	Links     bool `default:"true" json:"links,omitempty"`
	Meta      bool `default:"true" json:"meta,omitempty"`
	OpenGraph bool `default:"true" json:"opengraph,omitempty"`
}

type SelectorConfig struct {
	Slug       string                `json:"slug,omitempty"`
	Debug      bool                  `default:"true" json:"debug,omitempty"`
	Required   bool                  `default:"true" json:"required,omitempty"`
	Selector   string                `default:"css" json:"selector,omitempty"`
	Items      string                `json:"items,omitempty"`
	Details    map[string]Extractors `json:"details,omitempty"`
	StrictMode bool                  `default:"false" json:"strict_mode,omitempty"`
	// Type     string                `json:"type,omitempty"`
}

func (e *Endpoint) extractCss(sel *goquery.Selection, fields map[string]Extractors) Result { //extract 1 result using this endpoints extractor map
	r := Result{}
	if e.Debug {
		pp.Println(fields)
	}
	for field, ext := range fields {
		if v := ext.execute(sel); v != "" {
			if field == "url" && !strings.HasPrefix(v, "http") {
				r[field] = strings.Trim(fmt.Sprintf("%s%s", e.BaseURL, v), " ")
			} else {
				r[field] = strings.Trim(v, " ")
			}
		} else { //else if e.Debug {
			// r[field] = ""
			logf("missing field: %s", field)
		}
	}
	return r
}

func (e *Endpoint) extractXpath(node *html.Node, fields map[string]Extractors) Result { //extract 1 result using this endpoints extractor map
	pp.Print(e)
	r := Result{}
	for field, ext := range fields {
		xpathRule := GetExtractorValue(ext)
		// logf("xpathRule: %s", xpathRule)
		if v := htmlquery.FindOne(node, xpathRule); v != nil {
			t := htmlquery.InnerText(v)
			// logf("field %s, InnerText: %s", field, t) // fmt.Printf("field: %s \n", field)
			switch field {
			case "url":
				url := htmlquery.SelectAttr(v, "href")
				if url == "" {
					return nil
				}
				if field == "url" && !strings.HasPrefix(url, "http") {
					r[field] = strings.Trim(fmt.Sprintf("%s%s", e.BaseURL, url), " ")
				} else {
					r[field] = strings.Trim(url, " ")
				}
			default:
				r[field] = strings.Trim(t, " ")
			}
		} else { //else if e.Debug {
			// r[field] = ""
			logf("missing field: %s", field)
		}
	}
	return r
}

func (e *Endpoint) Execute(params map[string]string) (map[string][]Result, error) { // Execute will execute an Endpoint with the given params
	pp.Print(e)
	url, err := template(true, fmt.Sprintf("%s%s", e.BaseURL, e.URL), params) //render url using params
	if err != nil {
		return nil, err
	}
	method := e.Method //default method
	if method == "" {
		method = "GET"
	}
	body := io.Reader(nil) //render body (if set)
	if e.Body != "" {
		s, err := template(true, e.Body, params)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(s)
	}
	req, err := http.NewRequest(method, url, body) //create HTTP request
	if err != nil {
		return nil, err
	}
	if e.Headers != nil {
		for k, v := range e.Headers {
			if e.Debug {
				logf("use header %s=%s", k, v)
			}
			req.Header.Set(k, v)
		}
	}
	resp, err := http.DefaultClient.Do(req) //make backend HTTP request
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if e.Debug { //show results
		logf("%s %s => %s", method, url, resp.Status)
	}

	aggregate := make(map[string][]Result, 0)

	// https://github.com/golang/go/wiki/Switch
	switch e.Selector {
	case "xpath":
		doc, err := htmlquery.Parse(resp.Body)
		if err != nil {
			return nil, err
		}
		for b, s := range e.Blocks {
			if s.Items != "" {
				pp.Print(s)
				var results []Result
				htmlquery.FindEach(doc, s.Items, func(i int, node *html.Node) {
					// pp.Print(node)
					r := e.extractXpath(node, s.Details)
					// r["id"] = strconv.Itoa(i)
					if len(r) == len(s.Details) && s.StrictMode {
						results = append(results, r)
					} else if len(r) > 0 && !s.StrictMode {
						results = append(results, r)
					}
					if r != nil {
						results = append(results, r)
					}
				})
				if results != nil {
					aggregate[b] = results
				}
			}
		}

	case "css":
		doc, err := goquery.NewDocumentFromReader(resp.Body) //parse HTML
		if err != nil {
			return nil, err
		}
		sel := doc.Selection

		for b, s := range e.Blocks {
			if e.Debug {
				pp.Print(b)
				pp.Print(s)
			}
			var results []Result
			if s.Items != "" {
				sels := sel.Find(s.Items)
				//if e.Debug {
				logf("list: %s => #%d elements", s.Items, sels.Length())
				//}
				sels.Each(func(i int, sel *goquery.Selection) {
					r := e.extractCss(sel, s.Details)
					if len(r) == len(s.Details) && e.StrictMode {
						results = append(results, r)
					} else if len(r) > 0 && !e.StrictMode {
						results = append(results, r)
					}
					// else if e.Debug {
					logf("excluded #%d: has %d fields, expected %d", i, len(r), len(s.Details))
					//}
				})
				/*
					g := goose.New()
					article := g.ExtractFromURL(results["url"])
					println("title", article.Title)
					println("description", article.MetaDescription)
					println("keywords", article.MetaKeywords)
					println("content", article.CleanedText)
					println("url", article.FinalURL)
					println("top image", article.TopImage)
				*/
			} else {
				results[0] = e.extractCss(sel, s.Details)
			}

			if results != nil {
				aggregate[b] = results
			}
		}
	default:
		fmt.Println("unkown selector type")
	}
	return aggregate, nil
}
