package scraper

/*
	Refs:
	- https://github.com/Financial-Times/vulcan-config-builder/blob/master/main.go
	- https://github.com/ruprict/vulcand-atd-transformer/blob/master/templates.go
	- https://github.com/vmattos/apps-registrator/blob/master/etcd/etcd.go
*/

type Service struct {
	Name              string
	HasHealthCheck    bool
	Addresses         map[string]string
	PathPrefixes      map[string]string
	PathHosts         map[string]string
	FailoverPredicate string
}

type vulcanConf struct {
	FrontEnds map[string]vulcanFrontend
	Backends  map[string]vulcanBackend
}

type vulcanFrontend struct {
	BackendID         string
	Route             string
	Type              string
	rewrite           vulcanRewrite
	FailoverPredicate string
}

type vulcanRewrite struct {
	ID         string
	Type       string
	Priority   int
	Middleware vulcanRewriteMw
}

type vulcanRewriteMw struct {
	Regexp      string
	Replacement string
}

type vulcanBackend struct {
	Servers map[string]vulcanServer
}

type vulcanServer struct {
	URL string
}
