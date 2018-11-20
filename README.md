[![GoDoc](https://godoc.org/github.com/linux4life798/lora6100?status.svg)](https://godoc.org/github.com/linux4life798/lora6100)
[![Build Status](https://travis-ci.org/linux4life798/lora6100.svg?branch=master)](https://travis-ci.org/linux4life798/lora6100)

# Description
This is an interface library for the [NiceRF Lora6100AES](http://www.nicerf.com/product_149_65.html).

# Notes
* This library assumes that the serial `RTS` line is connected to the `SET`
  pin on the Lora6100 module. This is used to flip the module into settings
  mode on demand.