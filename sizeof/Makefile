deps:
	GO111MODULE=on go mod vendor

bench:
	go test -bench=Benchmark -v -benchmem  ./.

test:
	go test -v -count=1 ./.