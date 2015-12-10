export GOPATH := $(CURDIR)/.build
export OUTDIR ?= $(CURDIR)/bin
export BINARY_NAME ?= cm

LINUX_ARCHES ?= amd64 ppc64 386
WINDOWS_ARCHES ?= amd64 386

all: build

build: prepare clean_outdir

	for goarch in $(LINUX_ARCHES); do \
		echo Building linux $$goarch ...; \
		cd $(GOPATH)/cmd && GOOS=linux GOARCH=$$goarch go build ${LDFLAGS} . ; \
		mv $(GOPATH)/cmd/cmd $(OUTDIR)/$(BINARY_NAME)-linux-$$goarch || exit 1 ; \
	done ; \

	for goarch in $(WINDOWS_ARCHES); do \
		echo Building windows $$goarch ...; \
		cd $(GOPATH)/cmd && GOOS=windows GOARCH=$$goarch go build ${LDFLAGS} . ; \
		mv $(GOPATH)/cmd/cmd.exe $(OUTDIR)/$(BINARY_NAME)-windows-$$goarch.exe || exit 1 ; \
	done ; \

	cp -av $(CURDIR)/etc $(OUTDIR)

prepare: clean
	mkdir -p $(GOPATH) $(GOPATH)/src $(OUTDIR)
	cp -av cmd $(GOPATH)
	mkdir -p $(GOPATH)/src/cm; for dir in log event receiver sender storage supervisor; do \
		cp -av $$dir $(GOPATH)/src/cm ; \
	done

clean:
	rm -rf $(GOPATH)

clean_outdir:
	rm -rf $(OUTDIR)/*

test: prepare
	cd $(GOPATH)/cmd && go build ${LDFLAGS} .
	mv $(GOPATH)/cmd/cmd $(OUTDIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)
	$(OUTDIR)/$(BINARY_NAME)-${GOOS}-${GOARCH} --config ./etc/example.ini --debug-http 127.0.0.1:6060
