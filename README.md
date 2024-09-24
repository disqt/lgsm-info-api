# lgsm-info-api

`lgsm-info-api` is a Go API that uses gamedig to return information about which game servers are running on the host.

It's an HTTP API making use of [Gin](https://github.com/gin-gonic/gin).

## Building

You can build the app by running:

```bash
go build
```

## Running

You can run the following to run the app on local:

```bash
go run ./cmd/main.go
```

The API will then be listening on `localhost:8080`.

## Testing

The API has one integration test located [here](./cmd/main_test.go).

## Deploying

The API is currently running through linux services, to ensure that it starts up on boot.

- To check whether the API is running, you can run:

```bash
systemctl status lgsm-info-api.service
```

- To stop the API, you can run:

```bash
systemctl stop lgsm-info-api.service
```

- To start the API, you can run:

```bash
systemctl start lgsm-info-api.service
```
