package scraper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	// "github.com/linkosmos/mapdecor"
	// "github.com/toukii/jsnm"
	// "github.com/byrnedo/mapcast"
	// "github.com/spf13/cast"
	"github.com/PuerkitoBio/goquery"
	// "github.com/roscopecoltran/css2xpath"
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

func contains(input []string, match string) bool {
	for _, value := range input {
		if value == match {
			return true
		}
	}
	return false
}

func dedup(input []string) []string {
	var output []string
	for _, value := range input {
		if !contains(output, value) {
			output = append(output, value)
		}
	}
	return output
}

func ParseJson(body io.ReadCloser) (jsonBody map[string]interface{}, err error) {
	bytes, err := ioutil.ReadAll(body)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(bytes, &jsonBody)

	return
}

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

func HasElem(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {

			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}

	return false
}

/*
func mapdecor() {
    input := map[string]interface{}{
      "key1": nil,
      "key2": "with",
      "val1": "1",
      "val2": "2",
      "val3": "3",
      "val4": "4",
    }

    decorFunc := func(input map[string]interface{}) (output map[string]interface{}) {
      partitonFunc := func(s string, i interface{}) bool {
        return strings.Contains(s, "val")
      }

      // For first (inputPartitioned[0]) partition we get key-values containing 'val' in key
      // For second (inputPartitioned[1]) partition what is left
      inputPartitioned := mapop.Partition(partitonFunc, input)

      // Assigning values key to first partition
      inputPartitioned[1]["values"] = inputPartitioned[0]

      return inputPartitioned[1]
    }


    got := Decorate(input, decorFunc)

  // Got
  // map[string]interface{}{
  //   "key1": nil,
  //   "key2": "with",
  //   "values": map[string]interface{}{
  //     "val1": "1",
  //     "val2": "2",
  //     "val3": "3",
  //     "val4": "4",
  //   }
  // }
}
*/

/*
// https://github.com/byrnedo/mapcast
type myStruct struct {
    Field int `json:"input_name" bson:"output_name"`
    Hidden float32 `json:"-" bson:"hidden_field"`
}

func mapcast() {
	myInput := map[string]string{"input_name": "345"}

	Cast(myInput, myStruct{})
	// returns map{"Field" : 345}

	CastViaJson(myInput, myStruct{})
	// returns map{"input_name" : 345}

	CastViaJsonToBson(myInput, myStruct{})
	// returns map{"output_name" : 345}

	myMultiInput := map[string][]string{"input_name" : []string{"345}}

	CastMultiViaJsonToBson(myMultiInput, myStruct{})
	// returns map{"output_name" : []int[345]}
}
*/

/*
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
*/
