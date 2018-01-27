#!/bin/sh

export ISBN_RANGE_FILE=RangeMessage.xml

[ -f coverage.out ] && rm coverage.out

echo ""
echo "### gocyclo:"
gocyclo *.go

echo ""
echo "### test:"
go test -coverprofile=coverage.out

echo ""
echo "### coverage:"
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o=coverage.html
