package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/qor/admin"
	"github.com/qor/qor"
	// "github.com/roscopecoltran/scraper/scraper"
	// "github.com/howcrazy/xconv"
	// "github.com/leighmcculloch/go-structmap"
	// "github.com/goadesign/gorma"
	// "github.com/sas1024/gorm-loggable"
	// "github.com/etcinit/ohmygorm"
	//
)

// Create a GORM-backend model
type Provider struct {
	gorm.Model
	Name string
}

// Create another GORM-backend model
// type Endpoints []Endpoint // `json:"list,omitempty"`
//	List []Endpoint `json:"list,omitempty"`
//}

func main() {

	DB, _ := gorm.Open("sqlite3", "admin.db")
	DB.AutoMigrate(&Provider{}, &Endpoint{}, &SelectorType{}, &ExtractorsConfig{}, &BlocksConfig{}, &HeaderConfig{}, &SelectorConfig{}, &Extractor{}, &ExtractConfig{})
	// Initalize
	Admin := admin.New(&qor.Config{DB: DB})
	// Create resources from GORM-backend model

	Admin.SetSiteName("Scraper Config")

	Admin.AddResource(&Provider{})
	endpoint := Admin.AddResource(&Endpoint{})
	endpoint.Meta(&admin.Meta{Name: "Selector", Config: &admin.SelectOneConfig{Collection: SelectorEngines, AllowBlank: false}})
	endpoint.Meta(&admin.Meta{Name: "Method", Config: &admin.SelectOneConfig{Collection: MethodTypes, AllowBlank: false}})

	// Search products with its name, code, category's name, brand's name
	endpoint.SearchAttrs("Name", "BaseURL", "Selector", "Provider.Name", "Disabled")

	/*
		endpointHeadersPropertiesRes := endpoint.Meta(&admin.Meta{Name: "Headers"}).Resource
		endpointHeadersPropertiesRes.NewAttrs(&admin.Section{
			Rows: [][]string{{"Name", "Value"}},
		})
		endpointHeadersPropertiesRes.EditAttrs(&admin.Section{
			Rows: [][]string{{"Name", "Value"}},
		})
	*/

	/*
		endpointBlocksPropertiesRes := endpoint.Meta(&admin.Meta{Name: "Blocks"}).Resource
		endpointBlocksPropertiesRes.NewAttrs(&admin.Section{
			Rows: [][]string{{"Name", "Value"}},
		})
		endpointBlocksPropertiesRes.EditAttrs(&admin.Section{
			Rows: [][]string{{"Name", "Value"}},
		})
	*/

	// selectors := Admin.AddResource(&SelectorType{})
	// selectors.Meta(&admin.Meta{Name: "Engine", Config: &admin.SelectOneConfig{Collection: SelectorEngines, AllowBlank: false}})

	// Admin.AddResource(&SelectorConfig{})
	// Register route
	mux := http.NewServeMux()
	// amount to /admin, so visit `/admin` to view the admin interface
	Admin.MountTo("/admin", mux)
	// Start GIN Server
	fmt.Println("Listening on: 8080")
	r := gin.Default()
	r.Any("/admin/*w", gin.WrapH(mux))
	r.Run()

}
