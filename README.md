# Embark Coding Assignment

#### Hamid Sajjadi

This repository includes the code for the Embark Coding assignment.
The assignment asked for the following:

> Create a simple in-memory cache with an HTTP interface
> Interface
>- HTTP POST /<key> with the value as UTF-8 body
>- HTTP GET /<key> replies with the value as body or 404 if no such key exists

## Build and Run

This project needs Go 1.22 as it is using the stdlib `net/http` package
and [it's routing enhancement](https://go.dev/blog/routing-enhancements).
If you have go 1.22 installed, you can simply use the following command to run the project:

```shell
go run .
```

This will run the server on port 8080 by default. You can change the port by setting the `PORT` environment variable.

If you don't have go 1.22 installed and you don't want to install it, you can use the provided `Dockerfile`
and `docker-compose.yml` config to run the project. Running the following command will build and run the project:

```shell
docker-compose up
```

This will run the in a docker, and the server will be available on port 8080.

To learn more about the available endpoints, go to the [API Documentation](#api-documentation) section.
To learn more about the configuration and what is configurable, go to the [Configuration](#configuration) section.

## API Documentation

Two endpoints are available in this project.

### `POST /{key}`:

This endpoint is used to set a value for a key. The body will be parsed as a plian UTF-8 string and will be stored in
the cache. If the body is empty, the server will return a 400 status code.

example:

```shell
curl --location 'localhost:8080/user1' \
--header 'Content-Type: text/plain' \
--data '1234'
```

### `GET /{key}`:

This endpoint is used to get the value of a key. If the key exists, the value will be returned as the response body.
If `{key}` does not exist in cache, the server will return a 404 status code.

example:

```shell
## GET /{key}
curl --location 'localhost:8080/user1'
```

## Configuration

The server is configurable using environment variables. You can include a `.env` file in the root of the project to set
the variables. The following table shows the available configuration keys and their descriptions:

| Key                  | Description                                                                                                                              | Required | Default Value     | Alternative Key                     |
|----------------------|------------------------------------------------------------------------------------------------------------------------------------------|----------|-------------------|-------------------------------------|
| SERVICE_NAME         | name of the service. If given a non-empty value, configuration Keys will be prefixed by it and you should use Alternative Keys for them. | No       | -                 | -                                   |
| PORT                 | port of web server                                                                                                                       | No       | 8080              | [SERVICE_NAME]_PORT                 |
| HOST                 | hostname of web server                                                                                                                   | No       | localhost         | [SERVICE_NAME]_HOST                 |
| DEBUG                | turns on or off debug mode. Will affect verbosity of logs                                                                                | No       | false             | [SERVICE_NAME]_DEBUG                |
| TTL_SECONDS          | Time to Live (TTL) of records of the cache in second                                                                                     | No       | 1800 (30 minutes) | [SERVICE_NAME]_CACHE_TTL_SECONDS    |
| EVICTION_INTERVAL_MS | Time between two cache eviction processes running in the background in milliseconds                                                      | No       | 1000 (1 second)   | [SERVICE_NAME]_EVICTION_INTERVAL_MS |

## Implementation

The code is seperated into multiple modules:

- Cache: includes the in-memory cache implementation
- Server: includes the HTTP server (router and handlers)
- Config: includes the configuration for the server
- Logger: creates a logger using [zerolog](https://github.com/rs/zerolog)

### Cache Module

The in-memory cache uses a simple map to store the key-value pairs.
The cache is thread-safe and uses a `RWMutex` to lock the map when reading or writing to it.
`RWMutex` ensures that locking when reading does not block other readers, but it blocks writers.

Also, the cache supports Time to Live (TTL) for each key-value pair. The TTL is set to 30 minutes by default.
To remove the expired keys, a background goroutine runs in an interval (which is configurable) and removes the expired
keys.



