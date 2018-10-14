#!/bin/bash

set -exo pipefail


function addDep() {
	pushd $1

	cat <<EOF >nothingtoreadhere.go
package $1

import (
    _ "git.example.com/toni/${2}.git"
)
EOF

	git add nothingtoreadhere.go
	git commit -m 'Add dependency'
	git push

	popd
}

function depEnsureUpdate() {
	pushd $1

	dep ensure -update

	git add Gopkg.*
	git commit -m 'Update dependencies set'
	git push

	popd
}

function delDep() {
	pushd $1

	cat <<EOF >nothingtoreadhere.go
package $1
EOF

	git add nothingtoreadhere.go
	git commit -m 'Delete dependency'
	git push

	popd
}

function addFile() {
	pushd $1

	cat <<EOF >almostnothingtoreadhere.go
package $1
EOF

	git add almostnothingtoreadhere.go
	git commit -m 'Add almost nothing'
	git push

	popd
}

function delFile() {
	pushd $1

	git rm almostnothingtoreadhere.go
	git commit -m 'Delete almost nothing'
	git push

	popd
}


cd

git config --global user.name 'T.O.N.I.'
git config --global user.email 'toni@localhost'

export GOPATH="$(pwd)/go"

mkdir -p go/src/git.example.com/toni
cd go/src/git.example.com/toni

sleep 1

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

while true; do
	for action in {add,del}File; do
		sleep 30; addDep {grumpy,lol}cat
		sleep 30; depEnsureUpdate grumpycat

		sleep 30; $action lolcat
		sleep 30; depEnsureUpdate grumpycat

		sleep 30; delDep grumpycat
		sleep 30; depEnsureUpdate grumpycat
	done
done
