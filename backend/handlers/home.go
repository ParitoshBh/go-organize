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
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
)

type fileObject struct {
	Name    string
	Path    string
	Size    string
	IsImage bool
}

type directoryObject struct {
	Name string
	Path string
}

type listingObject struct {
	Files       []fileObject
	Directories []directoryObject
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
	baseBucket := connections.GetBaseBucket()

	// ViewData variables for view vars
	type ViewData struct {
		CurrentPath       string
		IsObjectListEmpty bool
		ListingObject     listingObject
		Pagination        paginator
		Breadcrumbs       []utils.Breadcrumb
		User              models.User
		FlashMessage      string
	}

	queryPath := r.URL.Query().Get("path")
	path := utils.GetCurrentPath(queryPath)

	viewData := ViewData{}

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

	viewData.ListingObject = objects
	if len(viewData.ListingObject.Directories) == 0 && len(viewData.ListingObject.Files) == 0 {
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
}

func listObjects(ctx context.Context, s3Client *s3.S3, nextContinuationToken string, previousContinuationToken string, baseBucket string, path string) (listingObject, paginator, error) {
	var (
		object     listingObject
		pagination paginator
	)

	if !objectExists(baseBucket, path) {
		return object, pagination, errors.New("folder doesn't exist")
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
		return object, pagination, err
	}

	// loop through all sub-directories first
	for _, commonPrefix := range output.CommonPrefixes {
		name := *commonPrefix.Prefix
		objectPath := name[0 : len(name)-1]

		object.Directories = append(object.Directories, directoryObject{
			Name: utils.BuildObjectName(objectPath),
			Path: fmt.Sprintf("?path=%s", objectPath),
		})
	}

	// loop through all files
	for _, content := range output.Contents {
		object.Files = append(object.Files, fileObject{
			Name:    utils.BuildObjectName(*content.Key),
			Path:    fmt.Sprintf("object/%s", *content.Key),
			Size:    humanize.Bytes(uint64(*content.Size)),
			IsImage: isImage(*content.Key),
		})
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

	return object, pagination, nil
}

func getUserData(ctx context.Context) (models.User, error) {
	sessionManager := utils.GetSessionManager()
	goOrmDB := connections.GetGoOrmDBConnection()
	user := models.User{}

	if !sessionManager.Exists(ctx, "userId") {
		return user, errors.New("session has expired")
	}

	// find user
	result := goOrmDB.Select([]string{"id", "first_name", "last_name"}).First(&user, sessionManager.Get(ctx, "userId"))
	if result.RowsAffected == 0 {
		return user, errors.New("unable to find user")
	}

	// load user's config
	err := goOrmDB.Model(&user).Association("UserConfig").Find(&user.UserConfig)
	if err != nil {
		return user, err
	}

	return user, nil
}

func isImage(name string) bool {
	extension := filepath.Ext(name)

	return strings.HasPrefix(mime.TypeByExtension(extension), "image/")
}
