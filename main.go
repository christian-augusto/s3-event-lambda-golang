package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

const TEMP_FILES_DIR = "/tmp"
const BUCKET_NAME = "camsb-bucket-2"
const AWS_REGION = "us-east-1"

type Event struct {
	A string `json:"a"`
}

func main() {
	lambda.Start(handler)
}

func handler(event Event) error {
	return writeFileAndUpload("")
}

func writeFileAndUpload(someContent string) error {
	var err error = nil
	fileName := uuid.NewString() + ".log"
	localFilePath := TEMP_FILES_DIR + "/" + fileName
	fileContent := "Hello World" + "\n" + someContent + "\n"
	remoteFilePath := "/test" + "/" + fileName

	fmt.Println("criando arquivo local ...")

	file, err := os.Create(localFilePath)

	if err != nil {
		return err
	}

	fmt.Println("arquivo local criado ...")

	fmt.Println("escrevendo no arquivo local ...")

	_, err = file.WriteString(fileContent)

	if err != nil {
		return err
	}

	fmt.Println("terminando de escrever no arquivo local ...")

	file.Close()

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(AWS_REGION),
		},
	)

	if err != nil {
		return err
	}

	storageService := NewS3StorageService(sess, BUCKET_NAME)

	err = storageService.UploadFile(localFilePath, remoteFilePath)

	return err
}

type S3StorageService struct {
	session    *session.Session
	bucketName string
}

func NewS3StorageService(sess *session.Session, bucketName string) *S3StorageService {
	return &S3StorageService{
		session:    sess,
		bucketName: bucketName,
	}
}

func (sds *S3StorageService) DownloadFile(remoteFilePath string) (string, error) {
	var err error = nil

	filePath := TEMP_FILES_DIR + "/" + uuid.NewString() + "-" + "download"

	file, err := os.Create(filePath)

	if err != nil {
		return "", err
	}

	defer file.Close()

	_, err = s3manager.NewDownloader(sds.session).Download(
		file,
		&s3.GetObjectInput{
			Bucket: aws.String(sds.bucketName),
			Key:    aws.String(remoteFilePath),
		},
	)

	return filePath, err
}

func (sds *S3StorageService) UploadFile(localFilePath string, remoteFilePath string) error {
	var err error = nil

	file, err := os.Open(localFilePath)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = s3.New(sds.session).PutObject(&s3.PutObjectInput{
		Bucket: aws.String(sds.bucketName),
		Key:    aws.String(remoteFilePath),
		Body:   file,
	})

	return err
}

func (sds *S3StorageService) GetFilename(filePath string) string {
	splitedFilepath := strings.Split(filePath, "/")

	return splitedFilepath[len(splitedFilepath)-1]
}
