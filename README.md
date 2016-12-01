# test-mysqld-docker

testing with docker mysqld for golang.

## DOWNLOAD

```
$ go get github.com/soh335/test-mysqld-docker
```

## USAGE

```go
mysqld, err := mysqltest.NewMysqld(nil)
if err != nil {
    log.Fatal(err.Error())
}
db, err := sql.Open("mysql", mysqld.DSN())
if err != nil {
    log.Fatal(err.Error())
}
if err := db.Ping(); err != nil {
    log.Fatal("ping failed")
}
```

## SEE ALSO

* https://github.com/lestrrat/go-test-mysqld
* https://github.com/punytan/p5-Test-Docker-MySQL

## LICENSE

* MIT
