run:
	@go run main.go $(CURDIR)/shared/conf.d/providers.list.json

build:
	@go build -o $(CURDIR)/dist/scraper-local main.go

dist: deps
	@gox -verbose -os="darwin linux" -arch="amd64" -output="$(CURDIR)/dist/scraper-{{.OS}}" .

deps:
	@go get -v -u github.com/Masterminds/glide
	@go get -v -u github.com/mitchellh/gox
	@go get -v -u github.com/moovweb/rubex