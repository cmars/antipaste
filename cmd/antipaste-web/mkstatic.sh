#!/bin/bash

if [ ! -d "static" ]; then
	echo "You're not running this from antipaste project root, are you?"
	exit 1
fi

if [ -x "$GOPATH/bin/go-bindata" ]; then
	tar czvf - static | $GOPATH/bin/go-bindata -p=antipaste -f=staticArchive > static.go
	exit 0
else
	echo "You need to install go-bindata:"
	echo "  go get github.com/jteeuwen/go-bindata"
	exit 1
fi
