run:
	@go run main.go $(CURDIR)/conf.d/providers.json

build:
	@go build -o $(CURDIR)/bin/scraper main.go

dist: deps
	@gox -verbose -os="darwin linux" -arch="amd64" -output="$(CURDIR)/dist/scraper-{{.OS}}" .

deps:
	@go get -v -u github.com/Masterminds/glide
	@go get -v -u github.com/mitchellh/gox
	@go get -v -u github.com/moovweb/rubex