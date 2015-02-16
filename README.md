ifmt (input formatter) is used to format JSON and XML from stdin

Install:

	go get github.com/james-maloney/ifmt
	cd $GOPATH/src/github.com/james-maloney/ifmt && make

JSON Usage:

	curl http://some-json-api.com | ifmt

XML Usage:

	curl http://some-json-api.com | ifmt -xml

