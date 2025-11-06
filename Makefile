BINARY=engine
IMAGE=registry.gitlab.com/wallet-gpay/orders-system
GOPATH:=$(shell go env GOPATH)
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
SHORT_CI=$(shell git rev-parse --short=8 HEAD)
GITLAB_URL=$(shell git config --get remote.origin.url | sed -e "s/\.git//")

telegram-failure:
	curl --fail -X POST --header "content-type: application/json" --data "{\"chat_id\": \"${TELE_CHAT_ID}\", \"text\": \"Order service FAILED to deploy with [Pipeline](${BUILD_URL})\ \", \"parse_mode\": \"MarkdownV2\"}" "https://api.telegram.org/bot${TELE_GITLABCI_TOKEN}/sendMessage"


telegram-success:
	curl --fail -X POST --header "content-type: application/json" --data "{\"chat_id\": \"${TELE_CHAT_ID}\", \"text\": \"Order System Jenkins has deployed success for commit [${SHORT_CI}](${GITLAB_URL}/-/commit/${SHORT_CI}) \", \"parse_mode\": \"MarkdownV2\"}" "https://api.telegram.org/bot${TELE_GITLABCI_TOKEN}/sendMessage"

engine:
	go build -o ${BINARY}  ./cmd/api/main.go

install-libproto:
	go get -u github.com/micro/micro/v2/cmd/protoc-gen-micro@master
	go install github.com/micro/micro/v2
	go get -u github.com/golang/protobuf/protoc-gen-go@v1.2

build:
	go build -o orders-system *.go

test:
	go test -mod vendor  -p 1 `go list ./... | grep -v /vendor/ | grep  -v /cmd/ | grep  -v /infrastructure/ | grep  -v /utils/`

docker:
	docker build . -t orders-system:latest

gen-mocks:
	cd ${ROOT_DIR}/domain/repositories && mockery --all
	cd ${ROOT_DIR}/proto/order_system && mockery --all
	cd ${ROOT_DIR}/proto/service_merchant_fee && mockery --all
	cd ${ROOT_DIR}/proto/service_promotion && mockery --all
	cd ${ROOT_DIR}/proto/service_transaction && mockery --all
	cd ${ROOT_DIR}/proto/service_user && mockery --all
	cd ${ROOT_DIR}/proto/service_card && mockery --all


gen-go2proto:
	cd ${ROOT_DIR}/domain/entities && go2proto -p ./ -f ${ROOT_DIR}/protobufs/order_system

gen-proto-macos:
	# gen user service
	rm -rf ./proto/service_user 1>/dev/null 2>/dev/null && protoc --proto_path=./protobufs --micro_out=./proto --go_out=./proto ./protobufs/service_user/*.proto

	# gen promotion
	rm -rf ./proto/service_promotion && protoc --proto_path=./protobufs --micro_out=./proto --go_out=./proto ./protobufs/service_promotion/*.proto

	# gen for service_transaction
	rm -rf ./proto/service_transaction 1>/dev/null 2>/dev/null && protoc --proto_path=./protobufs --micro_out=./proto --go_out=./proto ./protobufs/service_transaction/*.proto 0>/dev/null

	# gen for service_statistic
	rm -rf ./proto/service_merchant_fee && protoc --proto_path=./protobufs --micro_out=./proto --go_out=./proto ./protobufs/service_merchant_fee/*.proto

	# gen for service_merchant_fee
	rm -rf ./proto/service_merchant_fee 1>/dev/null 2>/dev/null && protoc --proto_path=./protobufs --micro_out=./proto --go_out=./proto ./protobufs/service_merchant_fee/*.proto 0>/dev/null

	rm -rf ./proto/order_system 1>/dev/null  2>/dev/null && protoc --proto_path=./protobufs --go_out=plugins=grpc:./proto ./protobufs/order_system/*.proto 0>/dev/null

	rm -rf ./proto/service_card 1>/dev/null  2>/dev/null && protoc  --proto_path=./protobufs --go_out=plugins=grpc:./proto ./protobufs/service_card/*.proto 0>/dev/null

