BUILT=`date -u '+%Y-%m-%d_%H:%M:%S'`
COMMIT=`git rev-parse HEAD`
VERSION=`git describe --tags|sed -e "s/\-/\./g"`
BRANCH=`git rev-parse --abbrev-ref HEAD`
LDFLAGS=-ldflags "-w -s -X main.VERSION=$(VERSION) -X main.BRANCH=$(BRANCH) \
-X main.COMMIT=$(COMMIT) -X main.BUILT=$(BUILT)"

.PHONY: build-ctl
build-ctl:
	go build $(LDFLAGS) -o lotusctl -trimpath cmd/lotusctl/main.go

.PHONY: build-srv
build-srv:
	go build -o lotussrv -trimpath cmd/lotussrv/main.go

.PHONY: install-ctl
install-ctl:
	make build-ctl && mv lotusctl /usr/local/bin

.PHONY: install-srv
install-srv:
	make build-srv && mv lotussrv /usr/local/bin

.PHONY: run
run:
	./lotussrv server run --http :8080 --meta ./meta
