.PHONY: build-ctl
build-ctl:
	go build -o lotusctl -trimpath cmd/lotusctl/main.go

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
