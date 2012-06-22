#!/bin/sh

export CGO_ENABLED=0
for os in windows linux darwin
do
	for arch in 386 amd64
	do
		mkdir -p $os"_"$arch
		GOOS=$os GOARCH=$arch go build -o $os"_"$arch/antipaste .
	done
done
