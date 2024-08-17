# Interline Transitland Server <!-- omit in toc -->

[![GoDoc](https://godoc.org/github.com/interline-io/transitland-server?status.svg)](https://godoc.org/github.com/interline-io/transitland-server) ![Go Report Card](https://goreportcard.com/badge/github.com/interline-io/transitland-server)

## Table of Contents <!-- omit in toc -->
<!-- to update use https://marketplace.visualstudio.com/items?itemName=yzhang.markdown-all-in-one -->
- [Installation](#installation)
- [Usage](#usage)
- [Usage as a web service](#usage-as-a-web-service)
- [Development](#development)
- [Licenses](#licenses)


## Installation

`cd cmd/tlserver && go install .`

## Usage

The resulting [`tlserver`](docs/cli/tlserver.md) binary includes several core commands from `transitland-lib` and adds the `server` command.

The main subcommands are:
* [tlserver server](docs/cli/tlserver_server.md)	 - Run transitland server
* [tlserver version](docs/cli/tlserver_version.md)	 - Program version and supported GTFS and GTFS-RT versions
* [tlserver fetch](docs/cli/tlserver_fetch.md)	 - Fetch GTFS data and create feed versions
* [tlserver import](docs/cli/tlserver_import.md)	 - Import feed versions
* [tlserver sync](docs/cli/tlserver_sync.md)	 - Sync DMFR files to database
* [tlserver unimport](docs/cli/tlserver_unimport.md)	 - Unimport feed versions
* [tlserver rebuild-stats](docs/cli/tlserver_rebuild-stats.md)	 - Rebuild statistics for feeds or specific feed versions

## Usage as a web service

To start the server with the REST API endpoints, GraphQL API endpoint, GraphQL explorer UI, and image generation endpoints:

```
tlserver server --dburl "postgres://your_host/your_database"
```

Alternatively, the database connection string can be specified using `TL_DATABASE_URL` environment variable. For local development environments, you will usually need to add `?sslmode=disable` to the connection string.

Open http://localhost:8080/ in your web browser to see the GraphQL browser, or use the endpoints at `/query` or `/rest/...`

The "example" server instance configured by the  `tlserver` command runs without authentication or authorization. Auth configuration is beyond the scope of this example command but can be added by configuring the server in your own package and adding HTTP middlewares to set user context and permissions data. You can use `cmd/tlserver/main.go` as an example to get started; it uses only public APIs from this package. (Earlier versions of `tlserver` included more built-in auth middlewares, but in our experience these are almost always custom per-installation, and were removed from this repo.) Additionally, this example server configuration exposes Go profiler endpoints on `/debug/pprof/...`. 

## Development

1. Install `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
2. On macOS, you will need the GNU timeout command: `brew install coreutils`
3. You will also need GDAL tools for importing reference data sets: `apt-get install gdal-bin`, or `brew install gdal-bin`, etc.
4. Check out `github.com/interline-io/transitland-lib` which contains the necessary schema and migrations.
5. Set `TL_TEST_SERVER_DATABASE_URL` to the connection string to a test database
   - e.g. `postgresql://localhost:5432/tlv2_test_server?sslmode=disable`
   - You must also set `PGHOST=localhost`, `PGDATABASE=tlv2_test_server`, etc., to match this url
   - This requirement may be removed in the future
1. Initialize test database schema: `transitland-lib/internal/schema/postgres/bootstrap.sh`
   - This will create the `tlv2_test_server` database in postgres
   - Will halt with an error (intentionally) if this database already exists
   - Runs golang-migrate on the migrations in `transitland-lib/internal/schema/postgres/migrations`
   - Unpacks and imports the Natural Earth datasets bundled with `transitland-lib`
2. Initialize test fixtures: `./testdata/test_setup.sh`
   - Builds and installs the `cmd/tlserver` command
   - Sets up test feeds contained in `testdata/server/server-test.dmfr.json`
   - Fetches and imports feeds contained in `testdata/external`
   - Creates additional fixtures defined in `testdata/test_supplement.pgsql`
   - Note that temporary files will be created in `testdata/tmp`; these are excluded in `.gitignore`
3. Optional: Set `TL_TEST_REDIS_URL` to run some GBFS tests
4. Optional: Set `TL_TEST_FGA_ENDPOINT` to a running [OpenFGA](https://github.com/openfga/openfga) server to run authorization tests
5. Run all tests with `go test -v ./...`

Test cases generally run within transactions; you do not need to regenerate the fixtures unless you are testing migrations or changes to data import functionality.
  
## Licenses

`transitland-server` is released under a "dual license" model:

- open-source for use by all under the [GPLv3](LICENSE) license
- also available under a flexible commercial license from [Interline](mailto:info@interline.io)


