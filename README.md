# CBSD RESTFull API example in golang

## Run

```shell script
go run ./main.go [ -l listen]
```

## Endpoints

### List bhyve domains

```shell script
curl [-s] [-i] http://127.0.0.1:8080/api/v1/blist
```

### Start (f11a) bhyve domain

```shell script
curl -i -X POST http://127.0.0.1:8080/api/v1/bstart/f111a
```

### Stop (f11a) bhyve domain

```shell script
curl -i -X POST http://127.0.0.1:8080/api/v1/bstop/f111a
```

### Remove (f11a) bhyve domain

```shell script
curl -i -X POST http://127.0.0.1:8080/api/v1/bremove/f111a
```

### Create new (f11a) bhyve domain
See *.json files for sample

```shell script
curl -X POST -H "Content-Type: application/json" -d @examples/bhyve_create_minimal.json http://127.0.0.1:8080/api/v1/bcreate/f111a
```

This is a just simple example. Contributing is welcome!
