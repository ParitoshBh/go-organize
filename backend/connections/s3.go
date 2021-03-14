package connections

import (
	"go-organizer/backend/logger"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/tidwall/gjson"
)

var s3Client *s3.S3
var baseBucket string

// BuildS3Connection builds and return connection to S3
func BuildS3Connection() {
	_logger := logger.Logger

	// load file
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		_logger.Error(err.Error())
	}

	// validate file as json
	if !gjson.Valid(string(file)) {
		_logger.Fatal("Invalid config file")
	}

	s3Config := gjson.Parse(string(file)).Get("connections.s3")

	// Initialize s3 client
	newSession, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(s3Config.Get("accessKey").String(), s3Config.Get("secret").String(), ""),
		Endpoint:         aws.String(s3Config.Get("endpoint").String()),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(s3Config.Get("disableSSL").Bool()),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		_logger.Fatal(err.Error())
	}

	s3Client = s3.New(newSession)
	baseBucket = s3Config.Get("bucket").String()
}

func GetS3Client() *s3.S3 {
	return s3Client
}

func GetBaseBucket() string {
	return baseBucket
}
