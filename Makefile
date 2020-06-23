.PHONY: build-ctl
build-ctl:
	go build -o lotusctl -trimpath cmd/lotusctl/main.go

.PHONY: build-srv
build-srv:
	go build -o lotussrv -trimpath cmd/lotussrv/main.go

.PHONY: run
run:
	./lotussrv server run --http :8080 --meta ./meta
