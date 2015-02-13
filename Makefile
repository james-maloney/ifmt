all:
	go get github.com/james-maloney/termstyle
	cd ${GOPATH}/src/github.com/james-maloney/termstyle && make
	go install
