all: test build

GIT_DESCRIBE=$(shell git describe --always)
GIT_DESCRIBE_LONG=$(shell git describe --always --long)
GIT_DESCRIBE_NO_V=$(shell git describe --always | sed 's/^v//g')

clean:
	rm -rf ./bin

test:
	go test ./...

_manpage:
	asciidoctor -b manpage -a version="$(GIT_DESCRIBE)" -D $(OUTPUT_MAN) dist/garden.adoc

manpage:
	OUTPUT_MAN=bin/ $(MAKE) _manpage

_build:
	go build -ldflags "-X 'main.Version=$(GIT_DESCRIBE)' -X 'main.VersionLong=$(GIT_DESCRIBE_LONG)' -X 'main.ModulesPath=$(MODULES_PATH)' -X 'main.ReportsPath=$(REPORTS_PATH)'" -o $(OUTPUT_FILE) ./cmd/garden

build:
	OUTPUT_FILE=bin/garden $(MAKE) _build

_macrelease:
	mkdir -p ./bin/release/darwin-universal

	lipo -create -output bin/release/darwin-universal/garden bin/release/darwin-amd64/garden bin/release/darwin-arm64/garden

	rm bin/release/bin/darwin-*
	cp bin/garden.1 bin/release/darwin-universal/garden.1
	cp -r garden-modules/modules bin/release/darwin-universal/modules
	tar -cvzf bin/release/bin/darwin-universal.tar.gz -C bin/release darwin-universal

_linuxrelease:
	ARCH=$(ARCH) VERSION=$(GIT_DESCRIBE_NO_V) GARDEN_BIN=bin/release/linux-$(ARCH)/garden GARDEN_MAN=bin/release/linux-$(ARCH)/garden.1 GARDEN_MODULES=bin/release/linux-$(ARCH)/modules/ nfpm pkg --config dist/nfpm.yaml --packager deb --target bin/release/bin
	ARCH=$(ARCH) VERSION=$(GIT_DESCRIBE_NO_V) GARDEN_BIN=bin/release/linux-$(ARCH)/garden GARDEN_MAN=bin/release/linux-$(ARCH)/garden.1 GARDEN_MODULES=bin/release/linux-$(ARCH)/modules/ nfpm pkg --config dist/nfpm.yaml --packager rpm --target bin/release/bin

_release:
	mkdir -p ./bin/release/$(OS)-$(ARCH)

	if [[ "$(OS)" == "windows" ]]; then\
		GOOS=$(OS) GOARCH=$(ARCH) OUTPUT_FILE=bin/release/$(OS)-$(ARCH)/garden.exe $(MAKE) _build;\
	else\
		MODULES_PATH=$(MODULES_PATH) REPORTS_PATH=$(REPORTS_PATH) GOOS=$(OS) GOARCH=$(ARCH) OUTPUT_FILE=bin/release/$(OS)-$(ARCH)/garden $(MAKE) _build;\
		cp bin/garden.1 bin/release/$(OS)-$(ARCH)/garden.1;\
	fi

	cp -r garden-modules/modules bin/release/$(OS)-$(ARCH)/modules

	tar -cvzf bin/release/bin/$(OS)-$(ARCH).tar.gz -C bin/release $(OS)-$(ARCH)

release: clean manpage
	mkdir -p ./bin/release/bin

	MODULES_PATH="/usr/local/Homebrew/share/garden/modules" REPORTS_PATH="/usr/local/Homebrew/share/garden/reports" OS=darwin ARCH=amd64 $(MAKE) _release
	MODULES_PATH="/opt/homebrew/share/garden/modules" REPORTS_PATH="/opt/homebrew/share/garden/reports" OS=darwin ARCH=arm64 $(MAKE) _release
	if [[ "$(shell uname -s)" == "Darwin" ]]; then\
		$(MAKE) _macrelease;\
	fi

	MODULES_PATH="/usr/share/garden/modules" REPORTS_PATH="/usr/share/garden/reports" OS=linux ARCH=amd64 $(MAKE) _release
	MODULES_PATH="/usr/share/garden/modules" REPORTS_PATH="/usr/share/garden/reports" OS=linux ARCH=arm64 $(MAKE) _release
	if $(shell which nfpm); then\
		ARCH=amd64 $(MAKE) _linuxrelease;\
		ARCH=arm64 $(MAKE) _linuxrelease;\
	fi

	OS=windows ARCH=amd64 $(MAKE) _release
	OS=windows ARCH=arm64 $(MAKE) _release
