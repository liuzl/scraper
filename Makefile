SWAGGER_UI_VERSION=3.3.2

run:
	@go run main.go $(CURDIR)/shared/conf.d/providers.json

build:
	@go build -o $(CURDIR)/dist/scraper-local main.go
	@echo "$ ./dist/scraper-local ./shared/conf.d/providers.json"
	@echo ""
	@./dist/scraper-local ./shared/conf.d/providers.list.json

dist:
	gox -verbose -os="darwin linux" -arch="amd64" -output="./dist/scraper-{{.OS}}" $(glide novendor)
	# gox -verbose -os="darwin linux" -arch="amd64" -output="./dist/scraper-{{.OS}}" $(glide novendor)

deps:
	@go get -v -u github.com/Masterminds/glide
	@go get -v -u github.com/mitchellh/gox
	@go get -v -u github.com/moovweb/rubex

compose:
	@docker-compose up --remove-orphans scraper

swagger-ui:	
	curl -L -o $(CURDIR)/swagger-ui-${SWAGGER_UI_VERSION}.tar.gz https://github.com/swagger-api/swagger-ui/archive/v$(SWAGGER_UI_VERSION).tar.gz
	tar zxf $(CURDIR)/swagger-ui-$(SWAGGER_UI_VERSION).tar.gz
	mv $(CURDIR)/swagger-ui-$(SWAGGER_UI_VERSION) $(CURDIR)/swaggerui
	rm -f $(CURDIR)/swagger-ui-$(SWAGGER_UI_VERSION).tar.gz