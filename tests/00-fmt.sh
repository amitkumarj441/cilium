#!/usr/bin/env sh

set -ex

if [ $# -eq 0 ]; then
	exit 0
fi

diff=`gofmt -d -l -s "$@"`

if [ -n "$diff" ]; then
	echo "Unformatted Go source code:"
	echo "$diff"
	exit 1
fi

exit 0
