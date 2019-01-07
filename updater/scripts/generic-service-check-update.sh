#!/bin/bash
# We hardcode the service and just care about the version
process_name=$1
version=$2
official_name=$3
service_github_url=$4
shift 4
arguments=$@
parsed_args=$(eval echo ${arguments})

echo "process name: ${process_name}"
echo "version: ${version}"
echo "official name: ${official_name}"
echo "service_github_url: ${service_github_url}"
binary="${GOPATH}/src/${service_github_url}/cmd/${process_name}"

binary_directory=${GOBIN}

check_if_different() {
    echo "cd into ${1}"
    cd $1
    exit_status=$?
    if [ ${exit_status} != 0 -a ${exit_status} != 1 ]; then
        echo "check script: no such file or directory"
        exit 2
    fi
    echo "check script: current directory is $(pwd)"
    go build
    exit_status=$?

    if [ ${exit_status} != 0 -a ${exit_status} != 1 ]; then
        echo "check script: failed building binary"
        exit 2
    fi

    if [[ -z $(diff ${GOBIN}/${2} ./${2}) ]]; then
        echo "check script: already up to date"
        exit 1
    fi

   echo "check script: new version"
   exit 0
}

echo "check script: fetching"
echo "go get -d -u ${service_github_url}"

# fetch new version
go get -d -u ${service_github_url}
exit_status=$?

if [ ${exit_status} != 0 -a ${exit_status} != 1 ]; then
    exit 2
fi

echo "check script: fetched"
echo "check script: checking if different"

check_if_different ${binary} ${process_name}