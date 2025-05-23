module github.com/interline-io/transitland-server

go 1.24.2

require (
	github.com/99designs/gqlgen v0.17.72
	github.com/Masterminds/squirrel v1.5.4
	github.com/aws/aws-sdk-go v1.49.6
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.14
	github.com/aws/aws-sdk-go-v2/service/location v1.44.2
	github.com/flopp/go-staticmaps v0.0.0-20220221183018-c226716bec53
	github.com/getkin/kin-openapi v0.127.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/graph-gophers/dataloader/v7 v7.1.0
	github.com/hypirion/go-filecache v0.0.0-20160810125507-e3e6ef6981f0
	github.com/interline-io/log v0.0.0-20250425230611-851ec713ec98
	github.com/interline-io/transitland-dbutil v0.0.0-20250506013203-964770296cd6
	github.com/interline-io/transitland-jobs v0.0.0-20250506013316-3a895dfc53e9
	github.com/interline-io/transitland-lib v1.1.3-0.20250523113013-c4fbc4295f1c
	github.com/interline-io/transitland-mw v0.0.0-20250506013255-4b0eba879d63
	github.com/jmoiron/sqlx v1.4.0
	github.com/rs/zerolog v1.34.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.10.0
	github.com/tidwall/gjson v1.18.0
	github.com/twpayne/go-geom v1.6.1
	github.com/twpayne/go-polyline v1.1.1
	github.com/vektah/gqlparser/v2 v2.5.26
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.6.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.8.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.0.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/PuerkitoBio/rehttp v1.3.0 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/auth0/go-auth0 v0.17.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.19 // indirect
	github.com/aws/smithy-go v1.22.3 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/flopp/go-coordsparser v0.0.0-20201115094714-8baaeb7062d5 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang-migrate/migrate/v4 v4.18.3 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
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
	github.com/jackc/pgx/v5 v5.7.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jellydator/ttlcache/v2 v2.11.1 // indirect
	github.com/jlaffaye/ftp v0.0.0-20220524001917-dfa1e758f3af // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/openfga/go-sdk v0.2.3 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/snabb/isoweek v1.0.1 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/tidwall/geoindex v1.7.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/rtree v1.10.0 // indirect
	github.com/tidwall/tinylru v1.2.1 // indirect
	github.com/tkrajina/gpxgo v1.1.2 // indirect
	github.com/twpayne/go-shapefile v0.0.6 // indirect
	github.com/urfave/cli/v2 v2.27.6 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/image v0.10.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/grpc v1.64.1 // indirect
	gopkg.in/dnaeon/go-vcr.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Fork to allow exporting x- extensions
replace github.com/getkin/kin-openapi => github.com/irees/kin-openapi v0.0.0-20240827112008-5f0d6c653b17

// replace github.com/interline-io/transitland-lib => /Users/irees/src/interline-io/transitland-lib
// replace github.com/interline-io/transitland-dbutil => /Users/irees/src/interline-io/transitland-dbutil
// replace github.com/interline-io/transitland-mw => /Users/irees/src/interline-io/transitland-mw
// replace github.com/interline-io/transitland-jobs => /Users/irees/src/interline-io/transitland-jobs
// replace github.com/interline-io/log => /Users/irees/src/interline-io/log

tool github.com/99designs/gqlgen
