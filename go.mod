module github.com/interline-io/transitland-server

go 1.16

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/Masterminds/squirrel v1.5.0
	github.com/aws/aws-sdk-go v1.38.54
	github.com/flopp/go-staticmaps v0.0.0-20210425143944-2e6e19a99c28
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible
	github.com/golang/geo v0.0.0-20210108004804-a63082ebfb66
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/hypirion/go-filecache v0.0.0-20160810125507-e3e6ef6981f0
	github.com/interline-io/transitland-lib v0.8.5-0.20210730235741-247315f022f3
	github.com/jmoiron/sqlx v1.3.1
	github.com/lib/pq v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.8.0
	github.com/twpayne/go-geom v1.3.6
	github.com/vektah/gqlparser/v2 v2.1.0
)

// replace github.com/interline-io/transitland-lib => /Users/irees/src/interline-io/transitland-lib
