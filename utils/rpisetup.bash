#!/bin/bash

PIN_SET=17
PIN_CS=18

GPIO=/sys/class/gpio

ensure_exported() {
	local pin=$1
	if [ ! -d "$GPIO/gpio$pin" ]; then
		echo $pin > $GPIO/export
	fi
}

ensure_exported $PIN_SET
ensure_exported $PIN_CS
sleep 0.5

echo out > $GPIO/gpio$PIN_SET/direction
echo out > $GPIO/gpio$PIN_CS/direction
sleep 0.5

if [ $# -gt 0 ]; then
	echo $1 > $GPIO/gpio$PIN_SET/value
	shift
else
	echo 1 > $GPIO/gpio$PIN_SET/value
fi

if [ $# -gt 0 ]; then
	echo $1 > $GPIO/gpio$PIN_CS/value
	shift
else
	echo 1 > $GPIO/gpio$PIN_CS/value
fi