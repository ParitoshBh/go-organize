package handlers

import (
	"fmt"
	"go-organizer/backend/connections"
	"go-organizer/backend/logger"
	"go-organizer/backend/utils"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
)

// GetObject return requested object
func GetObject(w http.ResponseWriter, r *http.Request) {
	s3Client := connections.GetS3Client()
	baseBucket := connections.GetBaseBucket()
	_logger := logger.Logger

	routerVars := mux.Vars(r)
	path := routerVars["path"]

	output, err := s3Client.GetObjectWithContext(r.Context(), &s3.GetObjectInput{
		Bucket: aws.String(baseBucket),
		Key:    aws.String(path),
	})
	if err != nil {
		_logger.Error(err.Error())
	}

	// update headers
	w.Header().Add("Content-Length", fmt.Sprintf("%d", *output.ContentLength))
	w.Header().Add("Content-Type", *output.ContentType)

	_, err = io.Copy(w, output.Body)
	if err != nil {
		// TODO take user to error page and/or show flash message on homepage
		_logger.Error(err.Error())
	}
}

// CreateObject uploads a file to bucket or creates a new object
func CreateObject(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}

	sessionManager := utils.GetSessionManager()
	s3Client := connections.GetS3Client()
	baseBucket := connections.GetBaseBucket()
	_logger := logger.Logger

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
		if objectExists(baseBucket, newBucketPath) {
			_logger.Error("Folder already exists")

			sessionManager.Put(r.Context(), "FlashMessage", "Folder already exists")

			http.Redirect(w, r, fmt.Sprintf("/?path=%s", bucketPath[:len(bucketPath)-1]), http.StatusSeeOther)
			return
		}

		_logger.Infof("Creating new bucket -> %s", newBucketPath)
		_, err := s3Client.PutObjectWithContext(r.Context(), &s3.PutObjectInput{
			Bucket: aws.String(baseBucket),
			Key:    aws.String(newBucketPath),
		})
		if err != nil {
			_logger.Error(err.Error())
			sessionManager.Put(r.Context(), "FlashMessage", err.Error())
		}

		http.Redirect(w, r, fmt.Sprintf("/?path=%s", newBucketPath[:len(newBucketPath)-1]), http.StatusSeeOther)
		return
	}

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
	if objectExists(baseBucket, filename) {
		utils.SendJSONResponse(w, http.StatusOK, response{
			Status:  false,
			Message: "File already exists",
		})
	} else {
		_, err := s3Client.PutObjectWithContext(r.Context(), &s3.PutObjectInput{
			Bucket: aws.String(baseBucket),
			Key:    aws.String(filename),
			Body:   file,
		})
		if err != nil {
			_logger.Error(err.Error())
			utils.SendJSONResponse(w, http.StatusOK, response{
				Status:  false,
				Message: err.Error(),
			})
		} else {
			_logger.Info("Uploaded")
			utils.SendJSONResponse(w, http.StatusOK, response{
				Status:  true,
				Message: "",
			})
		}
	}
}

func objectExists(baseBucket string, filename string) bool {
	s3Client := connections.GetS3Client()
	bucketHeadRes := true
	_logger := logger.Logger

	_, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(baseBucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		bucketHeadRes = false
	}

	output, err := s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:  aws.String(baseBucket),
		Prefix:  aws.String(filename),
		MaxKeys: aws.Int64(1),
	})
	if err != nil {
		_logger.Error(err.Error())
		return true
	}

	if !bucketHeadRes && *output.KeyCount == 0 {
		return false
	}

	return true
}

// DeleteObject deletes a object in a bucket or a bucket
func DeleteObject(w http.ResponseWriter, r *http.Request) {
	_logger := logger.Logger

	baseBucket := connections.GetBaseBucket()
	s3Client := connections.GetS3Client()

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

		err := s3Client.ListObjectsV2PagesWithContext(r.Context(), &s3.ListObjectsV2Input{
			Bucket:  aws.String(baseBucket),
			Prefix:  aws.String(prefix),
			MaxKeys: aws.Int64(2),
		}, func(lovo *s3.ListObjectsV2Output, b bool) bool {
			deleteKeys := []*s3.ObjectIdentifier{}
			for _, content := range lovo.Contents {
				deleteKeys = append(deleteKeys, &s3.ObjectIdentifier{
					Key: content.Key,
				})
			}

			_, err = s3Client.DeleteObjectsWithContext(r.Context(), &s3.DeleteObjectsInput{
				Bucket: aws.String(baseBucket),
				Delete: &s3.Delete{
					Objects: deleteKeys,
				},
			})
			if err != nil {
				_logger.Error(err.Error())
				return false
			}

			return true
		})
		if err != nil {
			_logger.Error(err.Error())
			deleted = false
		}

		_, err = s3Client.DeleteObjectWithContext(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(baseBucket),
			Key:    aws.String(prefix),
		})
		if err != nil {
			_logger.Error(err.Error())
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
		_, err := s3Client.DeleteObjectWithContext(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(baseBucket),
			Key:    aws.String(objectName),
		})
		if err != nil {
			_logger.Errorf(err.Error())
		}

		_logger.Infof("Deleted object -> %s", objectName)
		_logger.Infof("Redirecting to -> %s", utils.GetReponseRedirect(currentBucketPath))

		http.Redirect(w, r, utils.GetReponseRedirect(currentBucketPath), http.StatusSeeOther)
	}
}
