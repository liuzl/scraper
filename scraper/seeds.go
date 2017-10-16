package scraper

// https://github.com/angeldm/optiqor/blob/master/db/seeds/product.go
// https://github.com/angeldm/optiqor/blob/master/config/admin/admin.go

var Seeds = struct {
	Topics []struct {
		Name string
	}
	Groups []struct {
		Name string
	}
}{}
