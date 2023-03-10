package main

import (
	"context"
	"encoding/json"
	"functions/models"
	"io"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func getData() (people models.PeopleInSpace) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = "us-east-1"
		return nil
	})

	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg)
	downloader := manager.NewDownloader(client)

	bucketName := os.Getenv("DATA_BUCKET")
	bucketKey := os.Getenv("BUCKET_KEY")

	s3Object, s3err := downloader.S3.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(bucketKey),
	})

	if s3err != nil {
		log.Println("Failed to retrieve file from S3", bucketName, bucketKey, s3err)
		panic(s3err)
	}

	log.Println("Successfully retrieved file from S3")

	defer s3Object.Body.Close()

	body, readErr := io.ReadAll(s3Object.Body)

	if readErr != nil {
		panic(readErr)
	}
	log.Println("Successfully read file from S3")

	fileJson := string(body)

	unmarshalErr := json.Unmarshal([]byte(fileJson), &people)

	if unmarshalErr != nil {
		panic(unmarshalErr)
	}

	log.Println("Successfully unmarshalled json from S3")

	return people
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	peopleInSpace := getData()

	if len(peopleInSpace.People) <= 0 {
		panic("No people in space?")
	}

	readJson, _ := json.Marshal(peopleInSpace)
	responseBody := string(readJson)

	return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,",
				"Access-Control-Allow-Methods": "GET",
			},
			Body: responseBody,
		},
		nil
}

func main() {
	runtime.Start(handleRequest)
}
