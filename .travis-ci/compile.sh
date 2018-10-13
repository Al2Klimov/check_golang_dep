#!/bin/bash

set -exo pipefail

function mkBin() {
	case "$2" in
		386)
			ENVS=($(eval 'echo GO386={387,sse2}'))
			;;
		arm)
			ENVS=($(eval 'echo GOARM={5,6,7}'))
			;;
		mips|mipsle)
			ENVS=($(eval 'echo GOMIPS={hard,soft}float'))
			;;
		*)
			ENVS=('')
			;;
	esac

	for e in "${ENVS[@]}"; do
		if [ -z "$e" ]; then
			SUF=''
		else
			SUF="-$(cut -d '=' -f 2 <<<"$e")"
		fi

		eval "GOOS=$1 GOARCH=$2 $e go build -o "'"$(basename "$(pwd)")"'".${1}-${2}$SUF ."
	done
}


go generate

# .../brew/Cellar/go/1.11.1/libexec/pkg/tool/darwin_amd64/link: running clang failed: exit status 1
# ld: unknown option: -z
# clang: error: linker command failed with exit code 1 (use -v to see invocation)
#mkBin android   arm

mkBin dragonfly amd64
mkBin solaris   amd64

# # github.com/Al2Klimov/check_golang_dep/vendor/github.com/golang/dep/internal/fs
# vendor/...: undefined: syscall.Errno
# vendor/...: undefined: syscall.EXDEV
# # github.com/Al2Klimov/check_golang_dep/vendor/github.com/nightlyone/lockfile
# vendor/...: undefined: isRunning
# # github.com/Al2Klimov/check_golang_dep/vendor/github.com/boltdb/bolt
# vendor/...: undefined: flock
# vendor/...: undefined: mmap
# vendor/...: undefined: munmap
# vendor/...: undefined: fdatasync
# vendor/...: undefined: funlock
#for a in 386 amd64; do mkBin plan9 "$a"; done

# # github.com/Al2Klimov/check_golang_dep
# ...: relocation target runtime.read_tls_fallback not defined
#for a in 386 amd64 arm{,64}; do mkBin darwin "$a"; done
for a in 386 amd64; do mkBin darwin "$a"; done

# # github.com/Al2Klimov/check_golang_dep/vendor/github.com/boltdb/bolt
# vendor/...: undefined: maxMapSize
#for a in 386 amd64 arm{,64} ppc64{,le} mips{,64}{,le} s390x; do mkBin linux "$a"; done
for a in 386 amd64 arm{,64} ppc64{,le} s390x; do mkBin linux "$a"; done

for o in freebsd netbsd openbsd; do
	for a in 386 amd64 arm; do mkBin "$o" "$a"; done
done
