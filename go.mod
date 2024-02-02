module github.com/interline-io/transitland-server

go 1.21.5

require (
	github.com/99designs/gqlgen v0.17.42
	github.com/Masterminds/squirrel v1.5.4
	github.com/aws/aws-sdk-go v1.44.218
	github.com/aws/aws-sdk-go-v2 v1.17.5
	github.com/aws/aws-sdk-go-v2/config v1.18.15
	github.com/aws/aws-sdk-go-v2/service/location v1.22.1
	github.com/flopp/go-staticmaps v0.0.0-20220221183018-c226716bec53
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/graph-gophers/dataloader/v7 v7.1.0
	github.com/hypirion/go-filecache v0.0.0-20160810125507-e3e6ef6981f0
	github.com/interline-io/log v0.0.0-20240126000327-05bb90e4de4f
	github.com/interline-io/transitland-dbutil v0.0.0-20240126000951-f2dcc062261d
	github.com/interline-io/transitland-lib v0.14.1-0.20240202032317-a962b2983040
	github.com/interline-io/transitland-mw v0.0.0-20240126001316-f41a91e24d87
	github.com/jellydator/ttlcache/v2 v2.11.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/lib/pq v1.10.7
	github.com/rs/zerolog v1.31.0
	github.com/stretchr/testify v1.8.4
	github.com/tidwall/gjson v1.17.0
	github.com/tidwall/rtree v1.10.0
	github.com/tidwall/tinylru v1.1.0
	github.com/twpayne/go-geom v1.5.1
	github.com/vektah/gqlparser/v2 v2.5.10
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.1.4 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.1.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.0.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v0.5.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.5.1 // indirect
	github.com/PuerkitoBio/rehttp v1.3.0 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/amberflo/metering-go/v2 v2.5.0 // indirect
	github.com/auth0/go-auth0 v0.17.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.15 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.26.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.5 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/digitalocean/go-workers2 v0.10.4 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/flopp/go-coordsparser v0.0.0-20201115094714-8baaeb7062d5 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/golang-jwt/jwt v3.2.1+incompatible // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.3 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jlaffaye/ftp v0.0.0-20220524001917-dfa1e758f3af // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
	github.com/openfga/go-sdk v0.2.3 // indirect
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/snabb/isoweek v1.0.1 // indirect
	github.com/sosodev/duration v1.2.0 // indirect
	github.com/tidwall/geoindex v1.7.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tkrajina/gpxgo v1.1.2 // indirect
	github.com/urfave/cli/v2 v2.27.1 // indirect
	github.com/xrash/smetrics v0.0.0-20231213231151-1d8dd44e695e // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/image v0.10.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/oauth2 v0.15.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.16.1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/grpc v1.59.0 // indirect
	gopkg.in/dnaeon/go-vcr.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// replace github.com/interline-io/transitland-lib => /Users/irees/src/interline-io/transitland-lib
// replace github.com/interline-io/transitland-mw => /Users/irees/src/interline-io/transitland-mw
// replace github.com/interline-io/log => /Users/irees/src/interline-io/log
// replace github.com/interline-io/transitland-dbutil => /Users/irees/src/interline-io/transitland-dbutil
