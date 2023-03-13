package main

import (
	"encoding/json"
	"net/http"
)

type ResponseMessage struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func (m ResponseMessage) AsJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "\t")
}

var (
	ErrMalformedPathJSON = Must(
		(ResponseMessage{
			Message:    ErrMalformedPath.Error(),
			StatusCode: http.StatusNotFound,
		}).AsJSON(),
	)
	ErrBucketRequiredJSON = Must(
		(ResponseMessage{
			Message:    ErrBucketRequired.Error(),
			StatusCode: http.StatusNotFound,
		}).AsJSON(),
	)
	ErrIterationErrorJSON = Must(
		(ResponseMessage{
			Message:    ErrIterationError.Error(),
			StatusCode: http.StatusInternalServerError,
		}).AsJSON(),
	)
	ErrMarshallingJSON = Must(
		(ResponseMessage{
			Message:    ErrMarshalling.Error(),
			StatusCode: http.StatusInternalServerError,
		}).AsJSON(),
	)
	ErrBucketNotFoundJSON = Must(
		(ResponseMessage{
			Message: ErrBucketNotFound.Error(),
			StatusCode: http.StatusNotFound,
		}).AsJSON(),
	)
	ErrObjectNotFoundJSON = Must(
		(ResponseMessage{
			Message: ErrObjectNotFound.Error(),
			StatusCode: http.StatusNotFound,
		}).AsJSON(),
	)
	ErrObjectIOResponseJSON = Must(
		(ResponseMessage{
			Message: ErrObjectIOResponse.Error(),
			StatusCode: http.StatusInternalServerError,
		}).AsJSON(),
	)
)
