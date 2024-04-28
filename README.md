# Vector aggregation service

This is a basic example of a web service that uses [gonum](https://www.gonum.org/) to compute statistics of JSON data.

## Usage

To build and launch the service, run:

```sh
docker compose up
```

and check that the service is live with

```sh
curl localhost:8080/status
```

which should return

```json
{ "status": "ok" }
```

Then use the service with:

```sh
curl -s -d @example.json localhost:8080/centroid | jq
```

to compute and pretty print a column-wise average vector of the [input example](./example.json) data.

## Develop

To enter a development shell with all dependencies available, run:

```sh
nix develop
```

and run tests with:

```sh
go test
```
