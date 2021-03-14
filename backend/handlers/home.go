package handlers

import (
	"context"
	"errors"
	"fmt"
	"go-organizer/backend/connections"
	"go-organizer/backend/logger"
	"go-organizer/backend/models"
	"go-organizer/backend/templmanager"
	"go-organizer/backend/utils"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
)

type object struct {
	Name        string
	Path        string
	Size        string
	IsDirectory bool
	IsImage     bool
}

type paginator struct {
	IsPaginated  bool
	NextToken    string
	NextPath     string
	PreviousPath string
}

// Home loads template for listing bucket objects
func Home(w http.ResponseWriter, r *http.Request) {
	_logger := logger.Logger
	sessionManager := utils.GetSessionManager()
	s3Client := connections.GetS3Client()

	// ViewData variables for view vars
	type ViewData struct {
		CurrentPath       string
		IsObjectListEmpty bool
		Objects           []object
		Pagination        paginator
		Breadcrumbs       []utils.Breadcrumb
		User              models.User
		FlashMessage      string
	}

	baseBucket := "go-organizer"
	queryPath := r.URL.Query().Get("path")
	path := utils.GetCurrentPath(queryPath)

	viewData := ViewData{}

	// objectInfo, err := s3Client.StatObject(ctx, baseBucket, queryPath, minio.StatObjectOptions{})
	// if err != nil {
	// 	_logger.Warn(err.Error())

	// use tkn from query param if paginating results
	nextContinuationToken := r.URL.Query().Get("ntkn")
	previousContinuationToken := r.URL.Query().Get("ptkn")

	// not an object - load object list
	objects, pagination, err := listObjects(r.Context(), s3Client, nextContinuationToken, previousContinuationToken, baseBucket, path)
	if err != nil {
		sessionManager.Put(r.Context(), "FlashMessage", err.Error())

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	viewData.Objects = objects
	if len(viewData.Objects) == 0 {
		viewData.IsObjectListEmpty = true
	} else {
		viewData.Pagination = pagination
	}

	// get user's data
	user, err := getUserData(r.Context())
	if err != nil {
		// TODO logout user, show error page and redirect
		_logger.Error(err.Error())
	} else {
		viewData.User = user
	}

	viewData.Breadcrumbs = utils.BuildBreadcrumbs(path)
	viewData.CurrentPath = path

	// set flash message, if available in session
	viewData.FlashMessage = sessionManager.PopString(r.Context(), "FlashMessage")

	templmanager.RenderTemplate(w, "home.html", viewData)
	// } else {
	// 	// requesting a resouce, don't render template
	// 	if objectInfo.Err != nil {
	// 		// TODO handle error
	// 		_logger.Warn(objectInfo.Err.Error())
	// 	}

	// 	object, err := minioClient.GetObject(ctx, baseBucket, queryPath, minio.GetObjectOptions{})
	// 	if err != nil {
	// 		_logger.Error(err.Error())
	// 	}

	// 	http.ServeContent(w, r, "somename.svg", time.Now(), object)
	// }
}

func listObjects(ctx context.Context, s3Client *s3.S3, nextContinuationToken string, previousContinuationToken string, baseBucket string, path string) ([]object, paginator, error) {
	var (
		objects    []object
		pagination paginator
	)
	// _logger := logger.Logger

	if !objectExists(baseBucket, path) {
		return objects, pagination, errors.New("Folder doesn't exist")
	}

	// input params
	params := s3.ListObjectsV2Input{
		Bucket:    aws.String(baseBucket),
		Prefix:    aws.String(path),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int64(30),
	}

	// use continuation token, if not empty string
	if nextContinuationToken != "" {
		params.SetContinuationToken(nextContinuationToken)
	}

	output, err := s3Client.ListObjectsV2(&params)
	if err != nil {
		return objects, pagination, err
	}

	// loop through all sub-directories first
	for _, commonPrefix := range output.CommonPrefixes {
		name := *commonPrefix.Prefix
		objectPath := name[0 : len(name)-1]

		objects = append(objects, object{
			Name:        utils.BuildObjectName(objectPath),
			Path:        fmt.Sprintf("?path=%s", objectPath),
			IsDirectory: true,
		})
	}

	// loop through all files
	for _, content := range output.Contents {
		objects = append(objects, object{
			Name: utils.BuildObjectName(*content.Key),
			// Path: fmt.Sprintf("?path=%s", *content.Key),
			Size: humanize.Bytes(uint64(*content.Size)),
		})
		// 		viewObject.IsImage = isImage(mObject.Key)
	}

	pagination = paginator{}
	if *output.IsTruncated || output.ContinuationToken != nil {
		pagination.IsPaginated = true
		if output.NextContinuationToken != nil {
			pagination.NextToken = *output.NextContinuationToken
			pagination.NextPath = fmt.Sprintf("%s?ntkn=%s", path, *output.NextContinuationToken)
		}

		if output.ContinuationToken != nil {
			if pagination.NextPath != "" {
				pagination.NextPath = fmt.Sprintf("%s&ptkn=%s", pagination.NextPath, *output.ContinuationToken)
			}

			if previousContinuationToken == "" {
				pagination.PreviousPath = "/"
			} else {
				pagination.PreviousPath = fmt.Sprintf("%s?ntkn=%s", path, previousContinuationToken)
			}
		}
	}

	return objects, pagination, nil
}

func getUserData(ctx context.Context) (models.User, error) {
	sessionManager := utils.GetSessionManager()
	goOrmDB := connections.GetGoOrmDBConnection()
	user := models.User{}

	if !sessionManager.Exists(ctx, "userId") {
		return user, errors.New("Session has expired")
	}

	result := goOrmDB.Select([]string{"first_name", "last_name"}).First(&user, sessionManager.Get(ctx, "userId"))
	if result.RowsAffected == 0 {
		return user, errors.New("Unable to find user")
	}

	return user, nil
}

// func isImage(name string) bool {
// 	extension := filepath.Ext(name)

// 	return strings.HasPrefix(mime.TypeByExtension(extension), "image/")
// }
