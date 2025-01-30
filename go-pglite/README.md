PostgreSQL Server

```
podman run -it --rm -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust postgres:latest
```

go-pglite
```
go run main.go
```

client
```
mysql -h 127.0.0.1 -P 4000 -u root
```
