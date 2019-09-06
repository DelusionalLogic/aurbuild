#!/bin/sh

sigint_handler()
{
	echo "* Killing $PID"
	kill $PID
	exit
}

trap sigint_handler SIGINT

while true; do
	echo "* Starting process"
	swagger serve --no-open -p 1080 -F swagger petstore-expanded.yaml &
	PID=$!
	inotifywait -e modify -e attrib petstore-expanded.yaml
	echo "* Killing $PID"
	kill $PID
done
