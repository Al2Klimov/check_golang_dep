#!/bin/bash

set -exo pipefail

cd

git config --global user.name 'T.O.N.I.'
git config --global user.email 'toni@localhost'

export GOPATH="$(pwd)/go"

mkdir -p go/src/git.example.com/toni
pushd go/src/git.example.com/toni

for cat in {lol,grumpy}cat; do
	while ! git clone http://git.example.com/toni/$cat.git; do
		sleep 1
	done

	pushd $cat

	cat <<EOF >nothingtoreadhere.go
package $cat
EOF

	dep init

	git add nothingtoreadhere.go Gopkg.*
	git commit -m 'Add nothing'
	git push -u origin master

	popd
done

popd
