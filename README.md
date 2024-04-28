# Go service with Nix

This is a simple Go service that uses Nix to manage its dependencies.

## Usage

To build, test and run the service, run:

```sh
nix run
```

And call the service with:

```sh
curl http://localhost:8080
```

## Develop

To enter a development shell with all dependencies available, run:

```sh
nix develop
```

And run tests with:

```sh
go test
```
