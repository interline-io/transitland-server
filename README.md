# Interline Transitland Server <!-- omit in toc -->

[![GoDoc](https://godoc.org/github.com/interline-io/transitland-server?status.svg)](https://godoc.org/github.com/interline-io/transitland-server) ![Go Report Card](https://goreportcard.com/badge/github.com/interline-io/transitland-server)

## Table of Contents <!-- omit in toc -->
<!-- to update use https://marketplace.visualstudio.com/items?itemName=yzhang.markdown-all-in-one -->
- [Installation](#installation)
- [Usage as a Web Service](#usage-as-a-web-service)
  - [`server` command](#server-command)
  - [Hasura](#hasura)
- [Licenses](#licenses)


## Installation

`cd cmd/transitland_server && go install .`


## Usage as a Web Service

`transitland-lib` can be used in a variety of ways to power a web service. Interline currently uses two approaches:

1. Populate a database with one or more feeds using `transitland-lib` and use the `transitland server` command to serve the Transitland v2 REST and/or v2 GraphQL API endpoints. These API endpoints are primarily read-only and focused on querying and analyzing transit data.

2. Populate a Postgres database with one or more feeds using `transitland-lib`, or just create an empty database using `transitland-lib`'s schema. Use [Hasura](https://hasura.io/) to provide a complete GraphQL API for reading and writing into the database. 

For more information about how these web services are used within the overall architecture of the Transitland platform, see https://www.transit.land/documentation#transitland-architecture 

### `server` command

To start the server with the REST API endpoints, GraphQL API endpoint, GraphQL explorer UI, image generation endpoints, and without authentication (all requests run as admin):

```
transitland server -playground -auth=admin -dburl=postgres://localhost/transitland_db?sslmode=disable
```

The above command assumes you have a Postgres database named "transitland_db" running on your machine under the default Postgres port. Alternatively set the database URL using the `TL_DATABASE_URL` environment variable.

Open `http://localhost:8080/` in your web browser or query it using an HTTP client.

To start stripped down, with only REST API endpoints (no GraphQL API or explorer UI, no static image API endpoints) and with JWT authentication:

```
transitland server -disable-image -disable-graphql -auth=jwt -jwt-audience="http://example.com" -jwt-issuer="<issuer_id>" -jwt-public-key-file="<key.pem>"
```

Note that if you are using `transitland server` with JWT authentication, place the process behind a reverse-proxy or gateway that can issue JWT tokens. At Interline, we use the combination of [Kong](https://docs.konghq.com/) and [Auth0](https://auth0.com/).

More options:

```
% transitland server --help
Usage: server
  -auth string

  -dburl string
    	Database URL (default: $TL_DATABASE_URL)
  -disable-graphql
    	Disable GraphQL endpoint
  -disable-image
    	Disable image generation
  -disable-rest
    	Disable REST endpoint
  -gtfsdir string
    	Directory to store GTFS files
  -jwt-audience string
    	JWT Audience
  -jwt-issuer string
    	JWT Issuer
  -jwt-public-key-file string
    	Path to JWT public key file
  -playground
    	Enable GraphQL playground
  -port string
    	 (default "8080")
  -s3 string
    	S3 bucket for GTFS files
  -timeout int
    	 (default 60)
  -validate-large-files
    	Allow validation of large files
```

### Hasura

[Hasura](https://hasura.io/) is a web service that can provide an "instant" GraphQL API based on a Postgres database and its schema. We combine Hasura with `transitland-lib` for projects that involve creating new or complex queries (since Hasura can be more flexible than the queries provided by `transitland server`) and projects that involve an API with full read and write access (for example, editing GTFS data, which is also not provided by `transitland server`). Note that Hasura's automatically generated database queries are not guaranteed to be efficient (on the other hand, `transitland server` is tuned to provide better performance).

To use Hasura with `transitland-lib` you can either import feeds into a new Postgres database (using the `transitland dmfr` command) or create a blank Postgres database (using the schema in `internal/schema/postgres.pgsql`). Configure Hasura to recognize all the tables and the foreign key relationships between them.

## Licenses

`transitland-server` is released under a "dual license" model:

- open-source for use by all under the [GPLv3](LICENSE) license
- also available under a flexible commercial license from [Interline](mailto:info@interline.io)

