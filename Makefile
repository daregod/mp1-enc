# Makefile
GO ?= go

.PHONY: go-build go-test go-vet go-fmt

go-build:
	$(GO) build -buildvcs=false -o mp1enc.exe ./cmd/mp1enc
	$(GO) build -buildvcs=false -o mp1info.exe ./cmd/mp1info

go-test:
	$(GO) test ./...

go-vet:
	$(GO) vet ./...

go-fmt:
	$(GO) fmt ./...