# gogettext

![Go](https://github.com/taylor-s-dean/gogettext/workflows/Go/badge.svg?branch=master)

## Overview

`gogettext` is a library that intends to bring many of the major features of gettext to Go. Its development was driven by the need to have a thread-safe localization library that uses the gettext Portable Object (.po) file format while providing access to the underlying message catalog to the consumer. The C gettext library was originally intended to be used in standalone programs rather than distributed environments. `gogettext` is designed to be free of environmental constraints.

Other gettext-like libraries exist for Go, but they either violate the requirement for thread safety or they do not provide safe access to the underlying message catalog. 

* [gotext](https://github.com/leonelquinteros/gotext) is the most similar library and should be considered as an alternative to `gogettext`. It is thread-safe and emulates much of the C gettext library functionality while freeing itself from many of the environmental constraints of gettext.
* [gettext](https://github.com/gosexy/gettext) provides Go bindings to the gettext library. This is very useful for standalone programs, but it has all of the limitations of the C gettext library.

## Documentation

See [Godoc](https://pkg.go.dev/github.com/taylor-s-dean/gogettext)
