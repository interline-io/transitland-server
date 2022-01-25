module github.com/interline-io/transitland-server

go 1.16

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/Masterminds/squirrel v1.5.0
	github.com/ReneKroon/ttlcache/v2 v2.11.0
	github.com/aws/aws-sdk-go v1.38.54
	github.com/dnaeon/go-vcr/v2 v2.0.1
	github.com/flopp/go-staticmaps v0.0.0-20210425143944-2e6e19a99c28
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible
	github.com/go-redis/redis/v8 v8.11.4
	github.com/golang/geo v0.0.0-20210108004804-a63082ebfb66
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/hypirion/go-filecache v0.0.0-20160810125507-e3e6ef6981f0
	github.com/interline-io/transitland-lib v0.8.9-0.20220120001539-4f0f05c4ec47
	github.com/jmoiron/sqlx v1.3.1
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.8.0
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.12.1
	github.com/tidwall/tinylru v1.1.0
	github.com/twpayne/go-geom v1.4.1
	github.com/vektah/gqlparser/v2 v2.1.0
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210105161348-2e78108cf5f8 // indirect
)

// replace github.com/interline-io/transitland-lib => /Users/irees/src/interline-io/transitland-lib
