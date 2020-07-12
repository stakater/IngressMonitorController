#!/bin/bash

function os::util::list_go_src_files() {
	find . -not \( \
		\( \
		-wholename './_output' \
		-o -wholename './.*' \
		-o -wholename '*/vendor/*' \
		\) -prune \
	\) -name '*.go' | sort -u
}


bad_files=$(os::util::list_go_src_files | xargs gofmt -s -l)
if [[ -n "${bad_files}" ]]; then
        echo "!!! gofmt needs to be run on the listed files"
        echo "${bad_files}"
        echo "Try running 'gofmt -s -w [path]'"
        exit 1
fi


