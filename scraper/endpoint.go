package scraper

import (
	"context"
	"crypto/md5"
	// "crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/xquery/html"
	"github.com/archivers-space/warc"
	"github.com/cnf/structhash"
	"github.com/gebv/typed"
	"github.com/go-resty/resty"
	"github.com/iancoleman/strcase"
	"github.com/jeevatkm/go-model"
	"github.com/k0kubun/pp"
	"github.com/karlseguin/cmap"
	"github.com/leebenson/conform"
	"github.com/mgbaozi/gomerge"
	"github.com/mmcdole/gofeed"
	"github.com/oleiade/reflections"
	"github.com/roscopecoltran/mxj"
	"github.com/tsak/concurrent-csv-writer"
	"golang.org/x/net/html"
	// "github.com/cevaris/ordered_map"
	// "github.com/iancoleman/orderedmap"
	// "golang.org/x/net/context/ctxhttp"
	// "github.com/datatogether/pdf"
	// "github.com/datatogether/linked_data"
	// "github.com/datatogether/linked_data/dcat"
	// "github.com/datatogether/linked_data/pod"
	// "github.com/datatogether/linked_data/sciencebase"
	// "github.com/datatogether/linked_data/jsonld"
	// "github.com/datatogether/linked_data/xmp"
	// "github.com/ctessum/requestcache"
	// "github.com/otiai10/cachely"
	// "github.com/buger/jsonparser"
	// "github.com/go-aah/aah"
	// "github.com/creack/spider"
	// "github.com/whyrusleeping/json-filter"
	// "github.com/wolfeidau/unflatten"
	// "github.com/jzaikovs/t"
	// "github.com/linkosmos/urlutils"
	// "github.com/microcosm-cc/bluemonday"
	// "github.com/kennygrant/sanitize"
	// "github.com/slotix/slugifyurl"
	// "github.com/antchfx/xpath"
	// "github.com/advancedlogic/GoOse"
	// "github.com/ynqa/word-embedding/builder"
	// "github.com/ynqa/word-embedding/config"
	// "github.com/ynqa/word-embedding/validate"
)

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
	- https://github.com/jpillora/scraper/commit/0b5e5ce320ffaaaf86fb3ba9cc49458df3406a86
	- https://github.com/KKRainbow/segmentation-server/blob/master/main.go
	- https://github.com/mhausenblas/github-api-fetcher/blob/master/main.go
	- https://github.com/hoop33/limo/blob/master/service/github.go#L39
	- https://github.com/creack/spider/blob/master/example_test.go
	- https://github.com/suwhs/go-goquery-utils/tree/master/pipes
	- https://github.com/andrewstuart/goq
*/

var cacheDuration = 3600 * time.Second

func csvWriterTest() {
	// Create `dump.csv` in `./shared/data` directory
	csv, err := ccsv.NewCsvWriter("./shared/data/dump.csv")
	if err != nil {
		panic("Could not open `sample.csv` for writing")
	}

	// Flush pending writes and close file upon exit of main()
	defer csv.Close()

	count := 99

	done := make(chan bool)

	for i := count; i > 0; i-- {
		go func(i int) {
			csv.Write([]string{strconv.Itoa(i), "bottles", "of", "beer"})
			done <- true
		}(i)
	}

	for i := 0; i < count; i++ {
		<-done
	}
}

func warcReadAllTest() {
	f, err := os.Open("./shared/testdata/test.warc")
	if err != nil {
		fmt.Println("error: ", err.Error())
		return
	}
	defer f.Close()

	records, err := warc.NewReader(f).ReadAll()
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	if len(records) <= 0 {
		fmt.Printf("record length mismatch: %d isn't enough records", len(records))
		return
	}

	for _, r := range records {
		fmt.Println(r.Type().String())
	}
}

func typedTest(path string) {
	// directly from a map[string]interace{}
	// typed := typed.New(a_map)
	// from a json []byte
	// typed, err := typed.Json(data)
	// from a file containing JSON
	typ, _ := typed.JsonFile(path)
	pp.Print(typ)
}

func simpleGet() {
	resp, err := resty.R().Get("http://httpbin.org/get") // GET request
	if err != nil {
		fmt.Println("error: ", err)
	}
	// explore response object
	fmt.Printf("\nError: %v", err)
	fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
	fmt.Printf("\nResponse Status: %v", resp.Status())
	fmt.Printf("\nResponse Time: %v", resp.Time())
	fmt.Printf("\nResponse Received At: %v", resp.ReceivedAt())
	fmt.Printf("\nResponse Body: %v", resp) // or resp.String() or string(resp.Body())
}

func goModel(req http.Request) {
	// let's say you have just decoded/unmarshalled your request body to struct object.
	tempPeople, _ := ParseJson(req.Body)
	people := People{}
	// tag your Product fields with appropriate options like
	// -, omitempty, notraverse to get desired result.
	// Not to worry, go-model does deep copy :)
	errs := model.Copy(&people, tempPeople)
	fmt.Println("Errors:", errs)

	fmt.Printf("\nSource: %#v\n", tempPeople)
	fmt.Printf("\nDestination: %#v\n", people)
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

// StarResult wraps a star and an error
type ScraperResult struct {
	List  map[string][]Result
	Error error
}

func (e *Endpoint) ExecuteParallel(ctx context.Context, params map[string]string, resChan chan<- *ScraperResult) { // Execute will execute an Endpoint with the given params

	currentPage, _ := strconv.Atoi(e.Pager["offset"])
	lastPage, _ := strconv.Atoi(e.Pager["max"])

	offsetHolder := e.Pager["offset_var"]
	params[offsetHolder] = e.Pager["offset"]

	limitHolder := e.Pager["limit_var"]
	params[limitHolder] = e.Pager["limit"]
	for k, v := range e.Parameters {
		if _, ok := params[k]; !ok {
			if e.Debug {
				fmt.Printf("[WARNING] Parameters missing: k=%s, v=%s \n", k, v)
			}
		}
	}
	if e.Debug {
		fmt.Println("params")
		pp.Println(params)
	}
	for currentPage <= lastPage {
		res, err := e.Execute(params)
		if err != nil {
			resChan <- &ScraperResult{
				Error: err,
				List:  nil,
			}
		} else {
			if len(res) == 0 {
				lastPage = currentPage
				break
			} else {
				resChan <- &ScraperResult{
					Error: err,
					List:  res,
				}
			}
		}
		//if e.Debug {
		fmt.Println("res count: ", len(res), ", currentPage: ", currentPage, ", lastPage: ", lastPage)
		//}
		currentPage++ // Go to the next page
		params[offsetHolder] = strconv.Itoa(currentPage)
	}
	fmt.Println("closing resChan...")
	close(resChan)
}

/*
func enhancedGet() {
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"page_no": "1",
			"limit":   "20",
			"sort":    "name",
			"order":   "asc",
			"random":  strconv.FormatInt(time.Now().Unix(), 10),
		}).
		SetHeader("Accept", "application/json").
		SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F").
		Get("/search_result")

	// Sample of using Request.SetQueryString method
	resp, err := resty.R().
		SetQueryString("productId=232&template=fresh-sample&cat=resty&source=google&kw=buy a lot more").
		SetHeader("Accept", "application/json").
		SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F").
		Get("/show_product")
}
*/

func (e *Endpoint) getCacheKey(req *http.Request, debug bool) (string, string, string, error) {
	reqBytes, err := httputil.DumpRequest(req, true)
	if err != nil {
		return "", "", "", errors.New("dump request")
	}
	if debug {
		pp.Println(string(reqBytes))
	}
	cacheKey := fmt.Sprintf("%s_%x-%s_%s", e.hash, md5.Sum(reqBytes), req.Method, req.URL.String())
	cacheSlug := slugifier.Slugify(cacheKey)
	if e.Debug {
		fmt.Println("cacheSlug: ", cacheSlug)
	}
	cacheFile := fmt.Sprintf("./shared/cache/internal/%s.json", cacheSlug)
	return cacheKey, cacheSlug, cacheFile, nil
}

func (e *Endpoint) getHash(crypto string) (string, error) { // Execute will execute an Endpoint with the given params
	/*
		hash, err := structhash.Hash(e, 1)
		if err != nil {
			return "", err
		}
		fmt.Println("hash: ", hash)
		fmt.Println(structhash.Version(hash))
		if crypto == "md5" {
			fmt.Printf("structhash.Md5: %x\n", structhash.Md5(e, 1))
			fmt.Printf(" md5.Sum: %x\n", md5.Sum(structhash.Dump(e, 1)))
		}
		if crypto == "sha1" {
			fmt.Printf("structhash.Sha1: %x\n", structhash.Sha1(e, 1))
			fmt.Printf("sha1.Sum: %x\n", sha1.Sum(structhash.Dump(e, 1)))
		}
	*/
	return fmt.Sprintf("%x", structhash.Sha1(e, 1)), nil
}

func (e *Endpoint) Execute(params map[string]string) (map[string][]Result, error) { // Execute will execute an Endpoint with the given params
	if e.Debug {
		fmt.Println("endpoint handler config: ")
		pp.Println(e)
	}
	url, err := template(true, fmt.Sprintf("%s%s", e.BaseURL, e.PatternURL), params) //render url using params
	if err != nil {
		return nil, err
	}

	if e.Debug {
		fmt.Println("url: ", url)
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

	isResty := false
	if isResty {
		// https://github.com/go-resty/resty#various-post-method-combinations
		restyResp, err := resty.R().Get(url)
		if e.Debug {
			// explore response object
			fmt.Printf("\nError: %v", err)
			fmt.Printf("\nResponse Status Code: %v", restyResp.StatusCode())
			fmt.Printf("\nResponse Status: %v", restyResp.Status())
			fmt.Printf("\nResponse Time: %v", restyResp.Time())
			fmt.Printf("\nResponse Received At: %v", restyResp.ReceivedAt())
			fmt.Printf("\nResponse Body: %v", restyResp) // or resp.String() or string(resp.Body())
		}
	}

	// if e.Cache {
	_, _, cacheFile, err := e.getCacheKey(req, e.Debug)
	if err != nil {
		return nil, err
	}
	//}

	isCacheExpired := cacheExpired(cacheFile, cacheDuration)
	if e.Debug && e.Cache {
		fmt.Printf("[ENDPOINT] isCacheExpired: %t\ncacheFile: %s \n", isCacheExpired, cacheFile)
	}
	if !isCacheExpired && e.Cache {
		return cacheContent(cacheFile)
	}

	resp, err := getClient().Do(req)
	if err != nil {
		pp.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	if e.Debug { //show received headers
		fmt.Println("Response Headers: ")
		pp.Println(resp.Header)
		fmt.Println("Response Headers to intercept: ")
		pp.Println(e.HeadersIntercept)
	}

	for k, v := range resp.Header {
		if contains(e.HeadersIntercept, k) {
			if e.Debug {
				logf(" [INTERCEPT] header key=%s, value=%s", k, v)
			}
		}
	}

	if resp.StatusCode != 200 {
		if e.Debug {
			logf("%s %s => %s", method, url, resp.Status)
		}
		if resp.StatusCode == 403 {
			return nil, errors.New("Unauthorized request")
		}
	}

	aggregate := make(map[string][]Result, 0)

	if e.Debug {
		fmt.Println("e.Selector: ", e.Selector)
	}

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
	// https://stackoverflow.com/questions/24879587/xml-newdecoderresp-body-decode-giving-eof-error-golang
	case "xml":
		mv, err := mxj.NewMapXmlReader(resp.Body)
		if err != nil {
			return nil, err
		}
		if e.Debug {
			pp.Print(mv)
		}
		if e.ExtractPaths {
			mxj.LeafUseDotNotation()
			if e.Debug {
				fmt.Println("mv.LeafPaths(): ")
				pp.Println(mv.LeafPaths())
			}
			e.LeafPaths = leafPathsPatterns(mv.LeafPaths())
			if e.Debug {
				for _, v := range e.LeafPaths {
					fmt.Println("path:", v) // , "value:", v.Value)
				}
			}
		}
		for b, s := range e.BlocksJSON {
			if s.Items != "" {
				r := e.extractMXJ(mv, s.Items, s.Details)
				if e.Debug {
					fmt.Println("extractMXJ: ")
					pp.Println(r)
				}
				if r != nil {
					aggregate[b] = r
				}
			}
		}
	case "json":
		var mv mxj.Map
		var err error
		mxj.JsonUseNumber = true
		if e.Collection {
			mv, err = mxj.NewMapJsonArrayReaderAll(resp.Body)
		} else {
			mv, err = mxj.NewMapJsonReaderAll(resp.Body)
		}
		if err != nil {
			if e.Debug {
				fmt.Println("NewMapJsonReaderAll: ", err)
			}
			return nil, err
		}
		if e.ExtractPaths {
			mxj.LeafUseDotNotation()
			e.LeafPaths = leafPathsPatterns(mv.LeafPaths())
			if e.Debug {
				fmt.Println("mv.LeafPaths(): ")
				pp.Println(mv.LeafPaths())
				for _, v := range e.LeafPaths {
					fmt.Println("path:", v)
				}
			}
		}
		for b, s := range e.BlocksJSON {
			if e.Debug {
				pp.Println("s.Items: ", s.Items)
				pp.Println("s.Details: ", s.Details)
			}
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
		xml := resp.Body
		feed, err := fp.Parse(xml)
		if err != nil {
			return nil, err
		}
		for b, s := range e.BlocksJSON {
			var results []Result
			if e.Debug {
				fmt.Println("[RSS] items count: ", len(feed.Items))
			}
			for _, item := range feed.Items {
				if item != nil {
					res := e.extractRss(item, s.Details)
					if len(res) > 0 {
						results = append(results, res)
					}
				}
			}
			/*
				// Additional info
				r["author"] = feed.Author
				r["categories"] = feed. Categories
				r["custom"] = feed.Custom
				r["copyright"] = feed.Copyright
				r["description"] = feed.Description
				r["type"] = feed.FeedType
				r["language"] = feed.Language
				r["title"] = feed.Title
				r["published"] = feed.Published
				r["updated"] = feed.Updated
			*/
			if len(results) > 0 {
				aggregate[b] = results
			}
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
				if e.Debug {
					logf("list: %s => #%d elements", s.Items, sels.Length())
				}
				sels.Each(func(i int, sel *goquery.Selection) {
					r := e.extractCss(sel, s.Details)
					if len(r) == len(s.Details) && e.StrictMode {
						results = append(results, r)
					} else if len(r) > 0 && !e.StrictMode {
						results = append(results, r)
					} else if e.Debug {
						logf("excluded #%d: has %d fields, expected %d", i, len(r), len(s.Details))
					}
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
				// results = append(results, e.extract(sel))
			}

			if results != nil {
				aggregate[b] = results
			}
		}
	default:
		fmt.Println("unkown selector type")
	}

	if len(aggregate) > 0 && e.Cache {
		err = cacheResponse(cacheFile, aggregate) // dump response
		if err != nil {
			return nil, err
		}
	}

	return aggregate, nil
}

func cacheExpired(cacheFile string, maxAge time.Duration) bool {
	fi, err := os.Stat(cacheFile)
	if err != nil {
		return true
	}
	expireTime := fi.ModTime().Add(maxAge)
	fmt.Println("maxAge: ", maxAge)
	fmt.Println("expireTime: ", expireTime)
	fmt.Println("expired: ", time.Now().After(expireTime))
	return time.Now().After(expireTime)
}

func cacheContentRaw(cacheFile string) ([]byte, error) {
	file, err := os.Open(cacheFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	return b, nil
}

func cacheContent(cacheFile string) (map[string][]Result, error) {
	aggregate := make(map[string][]Result, 0)
	file, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		return nil, err
	}
	json.Unmarshal(file, &aggregate)
	return aggregate, nil
}

func cacheResponse(cacheSlug string, aggregate map[string][]Result) error {
	dump, err := json.Marshal(aggregate)
	// dump, err := json.MarshalIndent(aggregate, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cacheSlug, dump, 0644)
	if err != nil {
		return err
	}
	return nil
}

func leafPathsPatterns(input []string) []string {
	var output []string
	var re = regexp.MustCompile(`.([0-9]+)`)
	for _, value := range input {
		value = re.ReplaceAllString(value, `[*]`)
		if !contains(output, value) {
			output = append(output, value)
		}
	}
	return dedup(output)
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
		} else if e.Debug {
			logf("missing field: %s", field)
		}
	}
	return r
}

func (e *Endpoint) extractRss(item *gofeed.Item, fields map[string]Extractors) Result { //extract 1 result using this endpoints extractor map
	var fieldsList []string
	if e.Debug {
		pp.Println(fields)
	}
	for k, v := range fields {
		if e.Debug {
			pp.Println("fieldsList: k=", k, ", v=", v[0].val)
		}
		fieldsList = append(fieldsList, strcase.ToCamel(v[0].val))
	}
	r := Result{}
	for _, field := range fieldsList {
		has, _ := reflections.HasField(item, field)
		if has {
			value, err := reflections.GetField(item, field)
			if err != nil {
				if e.Debug {
					fmt.Println("error: ", err)
				}
				continue
			}
			key := strings.ToLower(field)
			if e.Debug {
				pp.Println("reflected value: ", value)
			}
			if value != nil {
				r[key] = value
			}
		}
	}
	if e.Debug {
		fmt.Println("item attr length:", len(r))
	}

	if e.Debug {
		fmt.Println("Results for RSS feed:")
		pp.Println(r)
	}

	return r
}

// import "github.com/jzaikovs/t"
func (e *Endpoint) extractMXJ(mv mxj.Map, items string, fields map[string]Extractors) []Result { //extract 1 result using this endpoints extractor map
	var r []Result
	if e.Debug {
		pp.Println(fields)
	}
	list, err := mv.ValuesForPath(items)
	if err != nil {
		fmt.Println("Error: ", err)
		// return nil
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
	if e.Debug {
		pp.Print(e)
	}
	r := Result{}
	for field, ext := range fields {
		xpathRule := GetExtractorValue(ext)
		if e.Debug {
			logf("xpathRule: %s", xpathRule)
		}
		if v := htmlquery.FindOne(node, xpathRule); v != nil {
			t := htmlquery.InnerText(v)
			if e.Debug {
				logf("field %s, InnerText: %s", field, t) // fmt.Printf("field: %s \n", field)
			}
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
		} else if e.Debug {
			logf("missing field: %s", field)
		}
	}
	return r
}
