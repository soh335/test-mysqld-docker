![](https://travis-ci.org/soh335/test-mysqld-docker.svg?branch=master)

# test-mysqld-docker

Testing with docker mysqld for golang. Support inside and outside of docker container.

## DOWNLOAD

```
$ go get github.com/soh335/test-mysqld-docker
```

## USAGE

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
defer cancel()
mysqld, err := mysqltest.NewMysqld(ctx, "mysql:latest")
if err != nil {
    log.Fatal(err.Error())
}
defer mysqld.Stop()
db, err := sql.Open("mysql", mysqld.DSN())
if err != nil {
    log.Fatal(err.Error())
}
if err := db.Ping(); err != nil {
    log.Fatal("ping failed")
}
```
### INSIDE DOCKER CONTAINER

Require docker command for finding parent conatiner network and ip address of mysql container is created. If you can allow to mount parent docker socket, add -v option like this ( ```-v /var/run/docker.sock:/var/run/docker.sock```).

## SEE ALSO

* https://github.com/lestrrat/go-test-mysqld
* https://github.com/punytan/p5-Test-Docker-MySQL

## LICENSE

* MIT
