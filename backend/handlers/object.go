package handlers

import (
	"fmt"
	"go-organizer/backend/connections"
	"go-organizer/backend/logger"
	"go-organizer/backend/utils"
	"net/http"

	"github.com/minio/minio-go/v7"
)

// CreateObject uploads a file to bucket or creates a new object
func CreateObject(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}

	_logger := logger.Logger
	minioClient := connections.MinioClient
	ctx := connections.Context

	baseBucket := "go-organizer"

	err := r.ParseForm()
	if err != nil {
		_logger.Fatalf(err.Error())
	}

	// Uppy sends values as content type form-data
	err = r.ParseMultipartForm(0)
	if err != nil {
		_logger.Errorf(err.Error())
	}

	bucketPath := r.FormValue("bucketPath")

	formValues := r.Form
	_, valueExists := formValues["bucket"]
	if valueExists {
		newBucketPath := ""
		if bucketPath == "" {
			newBucketPath = fmt.Sprintf("%s/", r.FormValue("bucket"))
		} else {
			newBucketPath = fmt.Sprintf("%s%s/", bucketPath, r.FormValue("bucket"))
		}

		// check if object with same name already exists
		if isDuplicateObject(baseBucket, newBucketPath) {
			// TODO Show flash message
			http.Redirect(w, r, fmt.Sprintf("/?path=%s", bucketPath[:len(bucketPath)-1]), http.StatusSeeOther)
		} else {
			_logger.Infof("Creating new bucket -> %s", newBucketPath)
			_, err = minioClient.PutObject(ctx, baseBucket, newBucketPath, nil, 0, minio.PutObjectOptions{})
			if err != nil {
				_logger.Warnf(err.Error())
				// TODO show flash message with error
			} else {
				http.Redirect(w, r, fmt.Sprintf("/?path=%s", newBucketPath[:len(newBucketPath)-1]), http.StatusSeeOther)
			}
		}
	} else {
		_logger.Infof("Upload objects")

		_logger.Info(formValues)
		file, fileHeader, err := r.FormFile("files[]")
		if err != nil {
			_logger.Errorf(err.Error())
		}
		_logger.Info(fileHeader.Header)
		_logger.Info(fileHeader.Size)

		filename := ""
		if bucketPath != "" {
			filename = fmt.Sprintf("%s%s", bucketPath, fileHeader.Filename)
		} else {
			filename = fileHeader.Filename
		}
		_logger.Infof(filename)

		// check if object with same name already exists
		if isDuplicateObject(baseBucket, filename) {
			utils.SendJSONResponse(w, http.StatusOK, response{
				Status:  false,
				Message: "File already exists",
			})
		} else {
			uploadInfo, err := minioClient.PutObject(ctx, baseBucket, filename, file, fileHeader.Size, minio.PutObjectOptions{})
			if err != nil {
				_logger.Errorf(err.Error())
				utils.SendJSONResponse(w, http.StatusOK, response{
					Status:  false,
					Message: err.Error(),
				})
			} else {
				_logger.Info(uploadInfo)
				utils.SendJSONResponse(w, http.StatusOK, response{
					Status:  true,
					Message: "",
				})
			}
		}
	}
}

func isDuplicateObject(baseBucket string, filename string) bool {
	minioClient := connections.MinioClient
	ctx := connections.Context
	_logger := logger.Logger

	_, err := minioClient.StatObject(ctx, baseBucket, filename, minio.StatObjectOptions{})
	if err != nil {
		_logger.Warnf(err.Error())
		return false
	}

	return true
}

// DeleteObject deletes a object in a bucket or a bucket
func DeleteObject(w http.ResponseWriter, r *http.Request) {
	_logger := logger.Logger

	baseBucket := "go-organizer"
	minioClient := connections.MinioClient
	ctx := connections.Context

	err := r.ParseForm()
	if err != nil {
		_logger.Errorf(err.Error())
	}

	currentBucketPath := r.FormValue("bucketPath")
	objectName := fmt.Sprintf("%s%s/", currentBucketPath, r.FormValue("objectName"))

	if r.FormValue("isDirectory") == "true" {
		deleted := true

		prefix := fmt.Sprintf("%s/", r.FormValue("objectName"))
		if currentBucketPath != "" {
			prefix = fmt.Sprintf("%s%s", currentBucketPath, prefix)
		}

		objectsCh := minioClient.ListObjects(ctx, baseBucket, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: false,
		})

		errorCh := minioClient.RemoveObjects(ctx, baseBucket, objectsCh, minio.RemoveObjectsOptions{})

		// log errors received from RemoveObjects API
		for e := range errorCh {
			_logger.Errorf("Failed to remove " + e.ObjectName + ", error: " + e.Err.Error())
			deleted = false
		}

		if deleted {
			_logger.Infof("Deleted object -> %s", prefix)
			_logger.Infof("Redirecting to -> %s", utils.GetReponseRedirect(currentBucketPath))

			http.Redirect(w, r, utils.GetReponseRedirect(currentBucketPath), http.StatusSeeOther)
		} else {
			_logger.Infof("Unable to delete object -> %s", prefix)
			_logger.Infof("Redirecting to -> %s", utils.GetReponseRedirect(currentBucketPath))

			http.Redirect(w, r, utils.GetReponseRedirect(currentBucketPath), http.StatusSeeOther)
		}
	} else {
		err = minioClient.RemoveObject(ctx, baseBucket, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			_logger.Errorf(err.Error())
		}

		_logger.Infof("Deleted object -> %s", objectName)
		_logger.Infof("Redirecting to -> %s", utils.GetReponseRedirect(currentBucketPath))

		http.Redirect(w, r, utils.GetReponseRedirect(currentBucketPath), http.StatusSeeOther)
	}
}
