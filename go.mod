module github.com/interline-io/transitland-server

go 1.23.0

require (
	github.com/99designs/gqlgen v0.17.49
	github.com/Masterminds/squirrel v1.5.4
	github.com/aws/aws-sdk-go v1.44.218
	github.com/aws/aws-sdk-go-v2 v1.17.5
	github.com/aws/aws-sdk-go-v2/config v1.18.15
	github.com/aws/aws-sdk-go-v2/service/location v1.22.1
	github.com/flopp/go-staticmaps v0.0.0-20220221183018-c226716bec53
	github.com/getkin/kin-openapi v0.127.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/graph-gophers/dataloader/v7 v7.1.0
	github.com/hypirion/go-filecache v0.0.0-20160810125507-e3e6ef6981f0
	github.com/interline-io/log v0.0.0-20240613202707-4e3adcc06d2d
	github.com/interline-io/transitland-dbutil v0.0.0-20240919224942-33fd018e3aff
	github.com/interline-io/transitland-lib v0.17.0-rc0.0.20240917232312-d3c262b7d8de
	github.com/interline-io/transitland-mw v0.0.0-20240916234429-9feee256a19e
	github.com/jellydator/ttlcache/v2 v2.11.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/rs/zerolog v1.33.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/gjson v1.17.0
	github.com/tidwall/rtree v1.10.0
	github.com/tidwall/tinylru v1.1.0
	github.com/twpayne/go-geom v1.5.1
	github.com/vektah/gqlparser/v2 v2.5.16
	google.golang.org/protobuf v1.34.1
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.6.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.8.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v0.5.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
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
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/digitalocean/go-workers2 v0.10.4 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/flopp/go-coordsparser v0.0.0-20201115094714-8baaeb7062d5 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/yaml v0.3.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jlaffaye/ftp v0.0.0-20220524001917-dfa1e758f3af // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/openfga/go-sdk v0.2.3 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/riverqueue/river v0.11.4 // indirect
	github.com/riverqueue/river/riverdriver v0.11.4 // indirect
	github.com/riverqueue/river/riverdriver/riverpgxv5 v0.11.4 // indirect
	github.com/riverqueue/river/rivershared v0.11.4 // indirect
	github.com/riverqueue/river/rivertype v0.11.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/snabb/isoweek v1.0.1 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/tidwall/geoindex v1.7.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tkrajina/gpxgo v1.1.2 // indirect
	github.com/twpayne/go-polyline v1.1.1 // indirect
	github.com/urfave/cli/v2 v2.27.4 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.uber.org/goleak v1.3.0 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/image v0.10.0 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/oauth2 v0.15.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/grpc v1.59.0 // indirect
	gopkg.in/dnaeon/go-vcr.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// replace github.com/interline-io/transitland-lib => /Users/irees/src/interline-io/transitland-lib
// replace github.com/interline-io/transitland-dbutil => /Users/irees/src/interline-io/transitland-dbutil
// replace github.com/interline-io/transitland-mw => /Users/irees/src/interline-io/transitland-mw
// replace github.com/interline-io/log => /Users/irees/src/interline-io/log
// replace github.com/getkin/kin-openapi => /Users/irees/src/other/kin-openapi
