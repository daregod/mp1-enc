# Makefile
GO ?= go
VERSION ?= 0.1.0
ifeq ($(OS),Windows_NT)
  EXE := .exe
else
  EXE :=
endif
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test vet fmt dist clean

build:
	$(GO) build -buildvcs=false -trimpath -ldflags "$(LDFLAGS)" -o mp1enc$(EXE) ./cmd/mp1enc
	$(GO) build -buildvcs=false -trimpath -ldflags "$(LDFLAGS)" -o mp1info$(EXE) ./cmd/mp1info

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

clean:
	$(GO) clean
	rm -f mp1enc mp1enc.exe mp1info mp1info.exe

dist:
	@mkdir -p dist
	GOOS=windows GOARCH=amd64 $(GO) build -buildvcs=false -trimpath -ldflags "$(LDFLAGS)" -o dist/mp1enc_windows_amd64.exe ./cmd/mp1enc
	GOOS=linux   GOARCH=amd64 $(GO) build -buildvcs=false -trimpath -ldflags "$(LDFLAGS)" -o dist/mp1enc_linux_amd64 ./cmd/mp1enc
	GOOS=darwin  GOARCH=amd64 $(GO) build -buildvcs=false -trimpath -ldflags "$(LDFLAGS)" -o dist/mp1enc_darwin_amd64 ./cmd/mp1enc
	GOOS=darwin  GOARCH=arm64 $(GO) build -buildvcs=false -trimpath -ldflags "$(LDFLAGS)" -o dist/mp1enc_darwin_arm64 ./cmd/mp1enc
	GOOS=linux   GOARCH=arm64 $(GO) build -buildvcs=false -trimpath -ldflags "$(LDFLAGS)" -o dist/mp1enc_linux_arm64 ./cmd/mp1enc
	GOOS=windows GOARCH=arm64 $(GO) build -buildvcs=false -trimpath -ldflags "$(LDFLAGS)" -o dist/mp1enc_windows_arm64.exe ./cmd/mp1enc
