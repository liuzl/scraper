# Sniperkit-Scraper - Docker stack
[To do]

## Intro
[WIP]

### Features
-

### Goals
1.
2.
3.

### Quick Start
```bash
go get -v github.com/roscopecoltran/scraper
cd $GOPATH/src/github.com/roscopecoltran/scraper
go run *.go ./shared/conf.d/providers.list.json
```

#### Make

##### (DEV) Scraper
```bash
make build
make run
```

##### (DIST) Scraper
```bash
make dist
```

#### Crane
```bash
go get -v -u github.com/michaelsauter/crane
```

##### (DIST) Scraper
```bash
crane up dist
```

##### (DEV) Scraper
```bash
crane up dev
```

#### Docker-Compose
MacOSX: 
```bash
brew install docker
brew install docker-compose
```

#### (DIST) Scraper + ETCD3 / E3CH 
Bootsrap:
```bash
docker-compose build --no-cache scraper
docker-compose up scraper
```

Examples:
```bash
open http://localhost:3000/bing?query=dlib (bing search endpoint)
open http://localhost:3000/admin (scraper admin)
```

#### (DEV) Scraper + ETCD3 / E3CH 
Bootsrap:
```bash
docker-compose build --no-cache scraper_dev
docker-compose up scraper_dev
```

#### (DEV) Scraper + ETCD3 / E3CH + ELK
Bootsrap:
```bash
docker-compose build --no-cache scraper_elk
docker-compose up scraper_elk
```

Examples:
```bash
open http://localhost:8086/ (e3ch)
open http://localhost:5601/ (kibana v5.x)
```

#### ETCD3 / E3CH 
Bootsrap:
```bash
docker-compose build --no-cache e3w_dev
docker-compose up e3w_dev
```

Examples:
```bash
open http://localhost:8086/ (e3ch)
```
