#!/bin/bash
# We hardcode the service and just care about the version
process_name=$1
version=$2
shift 2
arguments=$@
parsed_args=$(eval echo ${arguments})

echo "process name: ${process_name}"

binary="${GOPATH}/src/github.com/watercompany/skywire-messaging/cmd/${process_name}"

service_github_url="github.com/watercompany/skywire-messaging"
binary_directory=${GOBIN}

build_and_copy() {
    cd $1
    go build

    cp ${2} ${GOBIN}/${2}
}

echo "fetching"
echo "go get -d -u ${service_github_url}"

# fetch new version
go get -d -u ${service_github_url}
exit_status=$?

if [ ${exit_status} != 0 -a ${exit_status} != 1 ]; then
    exit 1
fi

echo "fetched"

echo "updating..."

build_and_copy ${binary} ${process_name}

echo "updated"
echo "restarting..."

cd ${GOBIN}
if [ !${process_name}.pid ]; then
    echo "no previous instance running or ${process_name}.pid doesn't exists"
else
    pkill -9 -F ${process_name}.pid
fi

nohup ./${process_name} ${parsed_args} > ${process_name}.log 2>&1 &echo $! > ${process_name}.pid &sleep 3
