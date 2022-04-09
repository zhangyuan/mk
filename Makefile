build: mk mk-server

.PHONY: clean
clean:
	rm -rf bin

.PHONY: mk
mk:
	go build -o bin/mk cmd/mk/main.go 

.PHONY: mk-server
mk-server:
	go build -o bin/mk-server cmd/mk-server/main.go
