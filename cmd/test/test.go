package main

import (
	"github.com/mainflux/mainflux/pkg/errors"
)

func main() {
	err0 := errors.New("0")
	err1 := errors.New("1")
	err2 := errors.New("2")
	errB := errors.Wrap(err1, err0)
	err := errors.Wrap(err2, errB)
	contains := errors.Contains(err, errB)
	println(contains)
}
