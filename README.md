1.install Golang on a Linux server

2.install Docker

3.install Go client for Docker API

`go get github.com/docker/docker/client`

4.install other Go package

`go get github.com/Workiva/go-datastructures`

5.launch Memcached in container

`sudo docker run --name memcached -d -p 11211:11211 zehwei/memcached memcached -m 30000 -t 16 -c 15600`

6.launch Totoro

`cd go/src/totoro/main/`

`go run main.go`

7.other commands

`docker update --cpuset-cpus 0,16 b59b2047cd2a`

`sudo perf stat -p 3937659`

`sudo perf stat -e cache-misses -p 2466013`

