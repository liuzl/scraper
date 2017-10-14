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

	"github.com/roscopecoltran/scraper/scraper"
)

// Create a GORM-backend model
type Provider struct {
	gorm.Model
	Name string
}

// Create another GORM-backend model
type Endpoint struct {
	gorm.Model
	Name string
}

func main() {

	DB, _ := gorm.Open("sqlite3", "data/demo.db")
	DB.AutoMigrate(&User{}, &Product{})
	// Initalize
	Admin := admin.New(&qor.Config{DB: DB})
	// Create resources from GORM-backend model
	Admin.AddResource(&User{})
	Admin.AddResource(&Product{})
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
