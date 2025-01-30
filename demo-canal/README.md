# MySQL 8.4
```
podman run --rm --env MYSQL_ALLOW_EMPTY_PASSWORD=1 --env MYSQL_ROOT_HOST='%' -p3307:3306 -it container-registry.oracle.com/mysql/community-server:8.4 --gtid-mode=ON --enforce-gtid-consistency=ON
```
