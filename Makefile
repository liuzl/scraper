run:
	@go run main.go $(CURDIR)/shared/conf.d/providers.list.json

build:
	@go build -o $(CURDIR)/dist/scraper-local main.go

dist:
	gox -verbose -os="darwin linux" -arch="amd64" -output="./dist/scraper-{{.OS}}" $(glide novendor)

deps:
	@go get -v -u github.com/Masterminds/glide
	@go get -v -u github.com/mitchellh/gox
	@go get -v -u github.com/moovweb/rubex