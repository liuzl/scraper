package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/moovweb/css2xpath"
	/*
		"github.com/rakanalh/goscrape"
		"github.com/rakanalh/goscrape/extract"
		"github.com/rakanalh/goscrape/processors"
	*//*
		"github.com/gorilla/css/scanner"
		"github.com/moovweb/gokogiri"
		"github.com/moovweb/gokogiri/html"
		"github.com/moovweb/gokogiri/xml"
		"github.com/moovweb/gokogiri/xpath"
	*/)

var templateRe = regexp.MustCompile(`\{\{\s*(\w+)\s*(:(\w+))?\s*\}\}`)

func template(isurl bool, str string, vars map[string]string) (out string, err error) {
	out = templateRe.ReplaceAllStringFunc(str, func(key string) string {
		m := templateRe.FindStringSubmatch(key)
		k := m[1]
		value, ok := vars[k]
		if !ok { //missing - apply defaults or error
			if m[3] != "" {
				value = m[3]
			} else {
				err = errors.New("Missing param: " + k)
			}
		}
		if isurl { //determine if we need to escape
			queryi := strings.Index(str, "?")
			keyi := strings.Index(str, key)
			if queryi != -1 && keyi > queryi {
				value = url.QueryEscape(value)
			}
		}
		return value
	})
	return
}

func checkSelector(s string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	doc, _ := goquery.NewDocumentFromReader(bytes.NewBufferString(`<html>
		<body>
			<h3>foo bar</h3>
		</body>
	</html>`))
	doc.Find(s)
	return
}

func jsonerr(err error) []byte {
	return []byte(`{"error":"` + err.Error() + `"}`)
}

func logf(format string, args ...interface{}) {
	log.Printf("[scraper] "+format, args...)
}

func xPathToCss(xpath []string, xtype string) []string {
	fmt.Printf("xpath type: %s \n", xtype)
	var result []string
	for _, css := range xpath {
		switch xtype {
		case "local":
			result = append(result, css2xpath.Convert(css, css2xpath.LOCAL))
		default:
			result = append(result, css2xpath.Convert(css, css2xpath.GLOBAL))
		}
	}
	return result
}
