# Set Env on powershell (7.x?)

```
$env:<new_env_name>=<value>
$env:GREENLIGHT_DB_DSN='postgres://greenlight:greenlight@localhost?sslmode=disable'
```
must use quote, not double quote(?)

# Postgre user & password
username : `greenlight`
password: `greenlight`

### Create role

For linux & windows
```psql
\c greenlight # set current db to greenlight
CREATE ROLE greenlight with login password 'admin';
```

# Migrate table

```
migrate create --ext=.sql --seq -dir=migrations create_movies_table
```

### Run Migration on windows

Working on cmder
```
 migrate --path=./migrations --database="postgres://greenlight:greenlight@localhost/greenlight?sslmode=disable" up
```

not yet working on powershell 7.x
```
migrate --path=./migrations -database=$GREELIGHT_DB_DSN up
```

# Installing dependendcies, go get or install

Latest go 1.18 default
- `go get` install global tools
- `go install` install per project

# go env

Go variables
```
GOPATH --> where YOUR project lives
```

# Env and Shit

```
GO111MODULE="on"
GOARCH="amd64"
GOBIN=""
GOCACHE="/home/ynrfin/.cache/go-build"
GOENV="/home/ynrfin/.config/go/env"
GOEXE=""
GOEXPERIMENT=""
GOFLAGS=""
GOHOSTARCH="amd64"
GOHOSTOS="linux"
GOINSECURE=""
GOMODCACHE="/home/ynrfin/go/pkg/mod"
GONOPROXY=""
GONOSUMDB=""
GOOS="linux"
GOPATH="/home/ynrfin/go"
GOPRIVATE=""
GOPROXY="https://proxy.golang.org,direct"
GOROOT="/usr/local/go"
GOSUMDB="sum.golang.org"
GOTMPDIR=""
GOTOOLDIR="/usr/local/go/pkg/tool/linux_amd64"
GOVCS=""
GOVERSION="go1.18.3"
GCCGO="gccgo"
GOAMD64="v1"
AR="ar"
CC="gcc"
CXX="g++"
CGO_ENABLED="1"
GOMOD="/dev/null"
GOWORK=""
CGO_CFLAGS="-g -O2"
CGO_CPPFLAGS=""
CGO_CXXFLAGS="-g -O2"
CGO_FFLAGS="-g -O2"
CGO_LDFLAGS="-g -O2"
PKG_CONFIG="pkg-config"
GOGCCFLAGS="-fPIC -m64 -pthread -fmessage-length=0 -fdebug-prefix-map=/tmp/go-build168036338=/tmp/go-build -gno-record-gcc-switches"
```


