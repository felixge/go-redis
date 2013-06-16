# go-redis

A redis implementation in Go.

Thanks to US Airways treating me to a 6 hour delay going from PHL -> BRU, I
found myself with a little extra time, so I decided to see if I could implement
a Go server implementing the redis protocol and GET/SET commands.

The result is kind of neat, and seems to perform almost as well as redis itself
without attempting any forms of optimizations yet.

## Running

```
go run main.go
```


## Performance

The numbers below should give you a rough idea, but don't take the numbers too
seriously as I've not done any extensive analysis here.

**go-redis:**

```
$ redis-benchmark -t GET,SET -n 100000
====== SET ======
  100000 requests completed in 1.18 seconds
  50 parallel clients
  3 bytes payload
  keep alive: 1

98.25% <= 1 milliseconds
99.99% <= 2 milliseconds
100.00% <= 2 milliseconds
84459.46 requests per second

====== GET ======
  100000 requests completed in 1.17 seconds
  50 parallel clients
  3 bytes payload
  keep alive: 1

98.33% <= 1 milliseconds
100.00% <= 1 milliseconds
85616.44 requests per second
```

**redis:**

```
$ redis-benchmark -t GET,SET -n 100000
====== SET ======
  100000 requests completed in 1.11 seconds
  50 parallel clients
  3 bytes payload
  keep alive: 1

99.95% <= 1 milliseconds
99.97% <= 2 milliseconds
100.00% <= 2 milliseconds
89847.26 requests per second

====== GET ======
  100000 requests completed in 1.12 seconds
  50 parallel clients
  3 bytes payload
  keep alive: 1

99.95% <= 1 milliseconds
100.00% <= 1 milliseconds
89206.06 requests per second
```
