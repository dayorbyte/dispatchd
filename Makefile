
PROTOC := protoc -I=${GOPATH}/src:${GOPATH}/src/github.com/gogo/protobuf/protobuf/:.

all: build

protoc_present:
	which protoc

deps:
	go get github.com/boltdb/bolt
	go get github.com/gogo/protobuf/gogoproto
	go get github.com/gogo/protobuf/proto
	go get github.com/gogo/protobuf/protoc-gen-gogo
	go get github.com/rcrowley/go-metrics
	go get github.com/streadway/amqp
	go get github.com/wadey/gocovmerge

gen_all: deps gen_pb gen_amqp

gen_pb: gen_amqp protoc_present
	$(PROTOC) --gogo_out=. amqp/*.proto
	$(PROTOC) --gogo_out=. gen/*.proto

gen_amqp:
	# TODO: this requires a virtualenv to be set up with Mako. I should
	# rewrite this in go to remove the dependency.
	python scripts/amqp_gen.py
	gofmt -w amqp/*generated*.go

build: deps gen_all
	go build -o dispatchd github.com/jeffjenkins/dispatchd/server

test: gen_all
	go test -cover github.com/jeffjenkins/dispatchd/...

full_coverage: test
	# Output: $$GOPATH/all.cover
	python scripts/cover.py

real_line_count:
	find . | grep '.go$$' | grep -v pb.go | grep -v generated | xargs cat | wc -l