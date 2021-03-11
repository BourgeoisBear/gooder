BINAME := gooder

debug:
	$(info "NOTE: watcher requires that GOPATH has been set")
	watcher

build:
	go get
	go build -v -o ${BINAME}

install:
	go install -v

uninstall:
	rm -f ${GOPATH}/bin/${BINAME}
