package handlers

import (
	"context"
	"errors"
	"go-organizer/backend/connections"
	"go-organizer/backend/logger"
	"go-organizer/backend/models"
	"go-organizer/backend/templmanager"
	"go-organizer/backend/utils"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/minio-go/v7"
)

type object struct {
	Name        string
	Path        string
	Size        string
	IsDirectory bool
	IsImage     bool
}

// Home loads template for listing bucket objects
func Home(w http.ResponseWriter, r *http.Request) {
	_logger := logger.Logger
	minioClient := connections.MinioClient
	ctx := connections.Context

	// ViewData variables for view vars
	type ViewData struct {
		CurrentPath       string
		IsObjectListEmpty bool
		Objects           []object
		Breadcrumbs       []utils.Breadcrumb
		User              models.User
	}

	baseBucket := "go-organizer"
	queryPath := r.URL.Query().Get("path")
	path := utils.GetCurrentPath(queryPath)

	viewData := ViewData{}

	objectInfo, err := minioClient.StatObject(ctx, baseBucket, queryPath, minio.StatObjectOptions{})
	if err != nil {
		_logger.Warn(err.Error())

		// not an object - load object list
		viewData.Objects = listObjects(ctx, minioClient, baseBucket, path)

		if len(viewData.Objects) == 0 {
			viewData.IsObjectListEmpty = true
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

		templmanager.RenderTemplate(w, "home.html", viewData)
	} else {
		// requesting a resouce, don't render template
		if objectInfo.Err != nil {
			// TODO handle error
			_logger.Warn(objectInfo.Err.Error())
		}

		object, err := minioClient.GetObject(ctx, baseBucket, queryPath, minio.GetObjectOptions{})
		if err != nil {
			_logger.Error(err.Error())
		}

		http.ServeContent(w, r, "somename.svg", time.Now(), object)
	}
}

func listObjects(ctx context.Context, minioClient *minio.Client, baseBucket string, path string) []object {
	var objects []object
	_logger := logger.Logger

	objectCh := minioClient.ListObjects(ctx, baseBucket, minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: false,
	})

	for mObject := range objectCh {
		if mObject.Err != nil {
			_logger.Warn(mObject.Err)
		}

		viewObject := object{}

		if mObject.Key[len(mObject.Key)-1:] == "/" {
			objectPath := mObject.Key[0 : len(mObject.Key)-1]

			viewObject.Name = utils.BuildObjectName(objectPath)
			viewObject.Path = "?path=" + objectPath
			viewObject.IsDirectory = true
			objects = append([]object{viewObject}, objects...)
		} else {
			viewObject.Name = utils.BuildObjectName(mObject.Key)
			viewObject.IsDirectory = false
			viewObject.Size = humanize.Bytes(uint64(mObject.Size))
			viewObject.Path = "?path=" + mObject.Key
			viewObject.IsImage = isImage(mObject.Key)
			objects = append(objects, viewObject)
		}
	}

	return objects
}

func getUserData(ctx context.Context) (models.User, error) {
	sessionManager := utils.GetSessionManager()
	goOrmDB := connections.GetGoOrmDBConnection()
	user := models.User{}

	result := goOrmDB.Select([]string{"first_name", "last_name"}).First(&user, sessionManager.Get(ctx, "userId"))
	if result.RowsAffected == 0 {
		return user, errors.New("Unable to find user")
	}

	return user, nil
}

func isImage(name string) bool {
	extension := filepath.Ext(name)

	return strings.HasPrefix(mime.TypeByExtension(extension), "image/")
}
