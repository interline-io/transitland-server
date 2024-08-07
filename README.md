# Interline Transitland Server <!-- omit in toc -->

[![GoDoc](https://godoc.org/github.com/interline-io/transitland-server?status.svg)](https://godoc.org/github.com/interline-io/transitland-server) ![Go Report Card](https://goreportcard.com/badge/github.com/interline-io/transitland-server)

## Table of Contents <!-- omit in toc -->
<!-- to update use https://marketplace.visualstudio.com/items?itemName=yzhang.markdown-all-in-one -->
- [Installation](#installation)
- [Usage as a Web Service](#usage-as-a-web-service)
  - [`server` command](#server-command)
- [Licenses](#licenses)


## Installation

`cd cmd/tlserver && go install .`


## Usage as a Web Service

### `server` command

To start the server with the REST API endpoints, GraphQL API endpoint, GraphQL explorer UI, image generation endpoints, and without authentication (all requests run as admin):

```
tlserver server --playground --auth=admin --dburl=postgres://localhost/transitland_db?sslmode=disable
```

The above command assumes you have a Postgres database named "transitland_db" running on your machine under the default Postgres port. Alternatively set the database URL using the `TL_DATABASE_URL` environment variable.

Open `http://localhost:8080/` in your web browser or query it using an HTTP client.

To start stripped down, with only REST API endpoints (no GraphQL API or explorer UI, no static image API endpoints) and with JWT authentication:

```
tlserver server --disable-image --disable-graphql --auth=jwt --jwt-audience="http://example.com" --jwt-issuer="<issuer_id>" --jwt-public-key-file="<key.pem>"
```

Note that if you are using JWT authentication, place the process behind a reverse-proxy or gateway that can issue JWT tokens. At Interline, we use the combination of [Kong](https://docs.konghq.com/) and [Auth0](https://auth0.com/).

## Licenses

`transitland-server` is released under a "dual license" model:

- open-source for use by all under the [GPLv3](LICENSE) license
- also available under a flexible commercial license from [Interline](mailto:info@interline.io)


