package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type BucketObject struct {
	Bucket string
	Object string
}

func (bo BucketObject) Path() string {
	return bo.Bucket + "/" + bo.Object
}

type BucketObjectHandlerFunc = func(w http.ResponseWriter, r *http.Request, bo *BucketObject)

func ObjectFromPath(path string, lg *log.Logger) (*BucketObject, error) {
	var bo BucketObject
	if len(path) == 0 {
		return nil, ErrMalformedPath
	}

	if path == "/" {
		return &bo, nil
	}

	parts := strings.Split(path[1:], "/")

	if len(parts) == 1 {
		bo.Bucket = parts[0]
		return &bo, nil
	}

	bo.Bucket = parts[0]
	bo.Object = strings.Join(parts[1:], "/")
	return &bo, nil

}

type StorageProxy struct {
	storage     *storage.Client
	logger      *log.Logger
	ListBuckets http.HandlerFunc
	ListObjects BucketObjectHandlerFunc
	GetObject   BucketObjectHandlerFunc
}

func New(client *storage.Client, logger *log.Logger) *StorageProxy {
	return &StorageProxy{
		storage:     client,
		logger:      logger,
		ListBuckets: DefaultListBuckets(client, logger),
		ListObjects: DefaultListObjects(client, logger),
		GetObject:   DefaultGetObject(client, logger),
	}
}

func (p *StorageProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the URL path
	path := r.URL.Path
	bo, err := ObjectFromPath(path, p.logger)
	if err != nil {
		ErrorWriter(
			ErrMalformedPathJSON,
			http.StatusNotFound,
		)(w, p.logger)
		return
	}

	switch {
	// This should never happen.
	case bo.Bucket == "" && bo.Object != "":
		ErrorWriter(
			ErrBucketRequiredJSON,
			http.StatusNotFound,
		)(w, p.logger)
	// List buckets
	case bo.Bucket == "" && bo.Object == "":
		p.ListBuckets(w, r)
	// List objects
	case bo.Bucket != "" && bo.Object == "":
		p.ListObjects(w, r, bo)
	// Get object
	case bo.Bucket != "" && bo.Object != "":
		p.GetObject(w, r, bo)
	}
}

func DefaultListBuckets(p *storage.Client, lg *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buckets := p.Buckets(r.Context(), ProjectID)
		bucketNames := make([]string, 0)
		for {
			attrs, err := buckets.Next()
			if err != nil {
				if err == iterator.Done {
					break
				} else {
					ErrorWriter(
						ErrIterationErrorJSON, http.StatusInternalServerError,
					)(w, lg)
				}
			}
			bucketNames = append(bucketNames, attrs.Name)
		}
		b, err := json.MarshalIndent(bucketNames, "", "\t")
		if err != nil {
			lg.Fatal(err)
		}
		w.Write(b)
	}
}

func DefaultListObjects(p *storage.Client, lg *log.Logger) BucketObjectHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, bo *BucketObject) {
		// Check if the bucket exists
		bucket := p.Bucket(bo.Bucket)
		_, err := bucket.Attrs(r.Context())
		if err != nil {
			ErrorWriter(
				ErrBucketNotFoundJSON, http.StatusNotFound,
			)(w, lg)
			return
		}
		objects := bucket.Objects(r.Context(), nil)
		objectNames := make([]string, 0)
		for {
			attrs, err := objects.Next()
			if err == iterator.Done {
				break
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(ErrIterationErrorJSON)
				return
			}
			objectNames = append(objectNames, attrs.Name)
		}
		b, err := json.MarshalIndent(objectNames, "", "\t")
		if err != nil {
			ErrorWriter(
				ErrMarshallingJSON, http.StatusInternalServerError,
			)(w, lg)
			return
		}
		w.Write(b)
	}
}

func DefaultGetObject(p *storage.Client, lg *log.Logger) BucketObjectHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, bo *BucketObject) {
		// Check if the bucket exists
		bucket := p.Bucket(bo.Bucket)
		_, err := bucket.Attrs(r.Context())
		if err != nil {
			ErrorWriter(
				ErrBucketNotFoundJSON, http.StatusNotFound,
			)(w, lg)
		}

		// Check if the object exists
		object := bucket.Object(bo.Object)
		_, err = object.Attrs(r.Context())
		if err != nil {
			ErrorWriter(
				ErrObjectNotFoundJSON, http.StatusNotFound,
			)(w, lg)
			return
		}
		// Stream the contents back
		or, err := object.NewReader(r.Context())
		if err != nil {
			ErrorWriter(
				ErrObjectIOResponseJSON, http.StatusInternalServerError,
			)(w, lg)
			return
		}
		_, err = io.Copy(w, or)
		if err != nil {
			ErrorWriter(
				ErrObjectIOResponseJSON, http.StatusInternalServerError,
			)(w, lg)
			return
		}
	}
}


