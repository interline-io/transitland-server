module github.com/interline-io/transitland-server

go 1.18

require (
	github.com/99designs/gqlgen v0.17.26
	github.com/Masterminds/squirrel v1.5.3
	github.com/amberflo/metering-go/v2 v2.0.1
	github.com/aws/aws-sdk-go v1.44.218
	github.com/aws/aws-sdk-go-v2 v1.17.5
	github.com/aws/aws-sdk-go-v2/config v1.18.15
	github.com/aws/aws-sdk-go-v2/service/location v1.22.1
	github.com/digitalocean/go-workers2 v0.10.3
	github.com/flopp/go-staticmaps v0.0.0-20220221183018-c226716bec53
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible
	github.com/go-chi/chi v1.5.4
	github.com/go-chi/chi/v5 v5.0.8
	github.com/go-chi/cors v1.2.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-redis/redismock/v8 v8.11.5
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/graph-gophers/dataloader/v7 v7.1.0
	github.com/hypirion/go-filecache v0.0.0-20160810125507-e3e6ef6981f0
	github.com/interline-io/transitland-lib v0.12.1-0.20230425205311-eacd3f704b0c
	github.com/jellydator/ttlcache/v2 v2.11.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/lib/pq v1.10.7
	github.com/prometheus/client_golang v1.14.0
	github.com/rs/zerolog v1.29.0
	github.com/stretchr/testify v1.8.2
	github.com/tidwall/gjson v1.14.4
	github.com/tidwall/tinylru v1.1.0
	github.com/twpayne/go-geom v1.5.1
	github.com/vektah/gqlparser/v2 v2.5.1
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c
	google.golang.org/protobuf v1.29.1
	gopkg.in/dnaeon/go-vcr.v2 v2.3.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.4.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.2.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.0.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.9.0 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.15 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.30.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.5 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/flopp/go-coordsparser v0.0.0-20201115094714-8baaeb7062d5 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.1 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jlaffaye/ftp v0.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/segmentio/backo-go v0.0.0-20200129164019-23eae7c10bd3 // indirect
	github.com/snabb/isoweek v1.0.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tkrajina/gpxgo v1.2.1 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/image v0.6.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/tools v0.7.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// replace github.com/interline-io/transitland-lib => /Users/irees/src/interline-io/transitland-lib
