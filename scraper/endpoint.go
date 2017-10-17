package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/roscopecoltran/mxj"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/xquery/html"
	"github.com/gebv/typed"
	"github.com/k0kubun/pp"
	"github.com/karlseguin/cmap"
	"github.com/leebenson/conform"
	"github.com/mgbaozi/gomerge"
	"golang.org/x/net/html"
	// "github.com/whyrusleeping/json-filter"
	// "github.com/wolfeidau/unflatten"
	// "github.com/jzaikovs/t"
	// "github.com/linkosmos/urlutils"
	// "github.com/microcosm-cc/bluemonday"
	// "github.com/kennygrant/sanitize"
	// "github.com/slotix/slugifyurl"
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

func typedTest(path string) {
	// directly from a map[string]interace{}
	// typed := typed.New(a_map)

	// from a json []byte
	// typed, err := typed.Json(data)

	// from a file containing JSON
	typ, _ := typed.JsonFile(path)
	pp.Print(typ)

}

func cmapTest() {
	m := cmap.New()
	m.Set("power", 9000)
	value, _ := m.Get("power")
	pp.Print(value)
	m.Delete("power")
	m.Len()
}

type People struct {
	Name  string `json:"name"`
	Sex   string `json:"sex"`
	Age   int    `json:"age"`
	Times int    `json:"times"`
}

// body as string
func gomergeTest(body []byte) {

	var tom People
	tom = People{
		Name:  "tom",
		Sex:   "male",
		Age:   18,
		Times: 1,
	}

	var request_data map[string]interface{}
	if err := json.Unmarshal(body, &request_data); err != nil {
		panic(err)
	}
	if err := gomerge.Merge(&tom, request_data); err != nil {
		panic(err)
	}
	result, _ := json.Marshal(tom)
	fmt.Println(result)
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

// import "github.com/jzaikovs/t"
func (e *Endpoint) extractMXJ(mv mxj.Map, items string, fields map[string]Extractors) []Result { //extract 1 result using this endpoints extractor map
	var r []Result
	if e.Debug {
		pp.Println(fields)
	}
	list, err := mv.ValuesForPath("items")
	if err != nil {
		fmt.Println("Error: ", err)
	}
	if e.Debug {
		pp.Println(list)
	}
	for i := 0; i < len(list); i++ {
		l := Result{}
		for attr, field := range fields {
			var keyPath string
			var node []interface{}
			if len(field) == 1 {
				keyPath = fmt.Sprintf("%#s[%#d].%#s", items, i, field[0].val)
				if e.Debug {
					fmt.Println("field[0].val=", field[0].val, "keyPath: ", keyPath)
				}
				node, _ = mv.ValuesForPath(keyPath)
			} else {
				w := make(map[string]interface{}, len(field))
				var merr error
				for _, whl := range field {
					var keyName string
					if strings.Contains(whl.val, "|") {
						keyParts := strings.Split(whl.val, "|")
						if e.Debug {
							pp.Println(keyParts)
						}
						keyName = keyParts[len(keyParts)-1]
						whl.val = keyParts[0]
						if e.Debug {
							fmt.Println("keyName alias: ", keyName)
						}
					} else {
						keyParts := strings.Split(whl.val, ".")
						if e.Debug {
							pp.Println(keyParts)
						}
						keyName = keyParts[len(keyParts)-1]
						if e.Debug {
							fmt.Println("keyName alias", keyName)
						}
					}
					keyPath = fmt.Sprintf("%#s[%#d].%#s", items, i, whl.val)
					if e.Debug {
						fmt.Println("keyName: ", keyName, ", whl.vall=", whl.val, "keyPath: ", keyPath)
					}
					node, merr = mv.ValuesForPath(keyPath)
					if merr != nil {
						fmt.Println("Error: ", merr)
					}
					if node != nil {
						if len(node) == 1 {
							w[keyName] = node[0]
						} else if len(node) > 1 {
							w[keyName] = node
						}
					}
				}
				if e.Debug {
					fmt.Println("subkeys whitelisted and mapped: ")
					pp.Println(w)
				}
				l[attr] = w
				continue
			}
			if len(node) == 1 {
				l[attr] = node[0]
			} else if len(node) > 1 {
				l[attr] = node
			}
		}
		r = append(r, l)
	}
	return r
}

func (e *Endpoint) extractXpath(node *html.Node, fields map[string]Extractors) Result { //extract 1 result using this endpoints extractor map
	// pp.Print(e)
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
	if e.Debug {
		pp.Print(e)
	}
	url, err := template(true, fmt.Sprintf("%s%s", e.BaseURL, e.PatternURL), params) //render url using params
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

	if e.HeadersJSON != nil {
		for k, v := range e.HeadersJSON {
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

	if e.Debug || resp.StatusCode != 200 { //show results
		logf("%s %s => %s", method, url, resp.Status)
	}

	aggregate := make(map[string][]Result, 0)

	switch e.Selector {
	case "wiki":
		if e.Debug {
			fmt.Println("Using 'WIKI' extractor")
		}
	case "md":
		if e.Debug {
			fmt.Println("Using 'MARKDOWN' extractor")
		}
	case "csv":
	case "tsv":
		if e.Debug {
			fmt.Printf("Using '%s-DELIMITED' extractor \n", e.Selector)
		}
	case "xml":
		mv, err := mxj.NewMapXmlReader(resp.Body)
		if err != nil {
			return nil, err
		}
		if e.Debug {
			pp.Print(mv)
		}
	case "json":
		mxj.JsonUseNumber = true
		mv, err := mxj.NewMapJsonReaderAll(resp.Body)
		if err != nil {
			return nil, err
		}
		for b, s := range e.BlocksJSON {
			if s.Items != "" {
				r := e.extractMXJ(mv, s.Items, s.Details)
				if e.Debug {
					pp.Println(r)
				}
				if r != nil {
					aggregate[b] = r
				}
			}
			if e.Debug {
				fmt.Println(" - block_key: ", b)
				pp.Println(s)
			}
		}
	case "rss":
		fp := gofeed.NewParser()
		feed, err := fp.Parse(resp.Body)
		if err != nil {
			return nil, err
		}
		for b, s := range e.BlocksJSON {
			var results []Result
			if results != nil {
				aggregate[b] = results
			}
			if e.Debug {
				fmt.Println("block_key: ", b)
				pp.Println(s)
			}
		}
		/*
			"items":       feed.Items,
			"author":      feed.Author,
			"categories":  feed.Categories,
			"custom":      feed.Custom,
			"copyright":   feed.Copyright,
			"description": feed.Description,
			"type":        feed.FeedType,
			"language":    feed.Language,
			"title":       feed.Title,
			"published":   feed.Published,
			"updated":     feed.Updated,
		*/
		if e.Debug {
			pp.Print(feed)
		}
	case "xpath":
		doc, err := htmlquery.Parse(resp.Body)
		if err != nil {
			return nil, err
		}
		for b, s := range e.BlocksJSON {
			if s.Items != "" {
				if e.Debug {
					pp.Print(s)
				}
				var results []Result
				/*
					rules, err := ConvertDetails(s.DetailsJSON)
					if err != nil {
						return nil, err
					}
				*/
				htmlquery.FindEach(doc, s.Items, func(i int, node *html.Node) {
					r := e.extractXpath(node, s.Details)
					if len(r) == len(s.Details) {
						results = append(results, r)
					} else if len(r) > 0 {
						if s.StrictMode == false {
							results = append(results, r)
						}
					}
					conform.Strings(r)
					if r != nil {
						r["id"] = strconv.Itoa(i)
						results = append(results, r)
					}
					if e.Debug {
						fmt.Print(" ---[ result: \n")
						pp.Print(r)
						fmt.Print(" ]---- \n")
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

		for b, s := range e.BlocksJSON {
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
