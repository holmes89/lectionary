# Lectionary API

I was inspired by this [project](https://github.com/marmanold) to build my own verse search and lectionary tool. My hope is to extend this into a service that
allows others to document their reflections on various passages.

## Running

- You must first build the database using `go run cmd/load/main.go`
- Then run `go run cmd/server/main.go`
- Searching by verse `curl localhost:8080/verse/?q=Joel+1:3-5`
