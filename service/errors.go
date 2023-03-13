package main

import (
	"errors"
	"log"
	"net/http"
)

var (
	ErrMalformedPath    error = errors.New("malformed object path")
	ErrBucketRequired   error = errors.New("bucket is required")
	ErrBucketNotFound   error = errors.New("bucket not found")
	ErrIterationError   error = errors.New("non-nil error during iteration")
	ErrMarshalling      error = errors.New("json marshalling error")
	ErrObjectNotFound   error = errors.New("object not found")
	ErrObjectIOResponse error = errors.New("object io error")
)

func ErrorWriter(err []byte, status int) func(http.ResponseWriter, *log.Logger) {
	return func(w http.ResponseWriter, lg *log.Logger) {
		lg.Println(string(err))
		w.WriteHeader(status)
		w.Write(err)
	}
}
