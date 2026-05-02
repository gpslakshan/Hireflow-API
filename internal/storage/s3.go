package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appconfig "github.com/gpslakshan/hireflow/internal/config"
)

// S3Storage wraps the AWS S3 pre-signed client.
type S3Storage struct {
	presignClient  *s3.PresignClient
	bucket         string
	uploadExpiry   time.Duration
	downloadExpiry time.Duration
}

// NewS3Storage initialises the S3 client from config.
func NewS3Storage(cfg *appconfig.Config) (*S3Storage, error) {
	uploadMins, err := strconv.Atoi(cfg.CVUploadURLExpiryMins)
	if err != nil {
		uploadMins = 15
	}
	downloadMins, err := strconv.Atoi(cfg.CVDownloadURLExpiryMins)
	if err != nil {
		downloadMins = 15
	}

	// Build AWS config with explicit credentials from .env
	awsCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AWSAccessKeyID,
				cfg.AWSSecretAccessKey,
				"", // session token — empty for IAM user keys
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)
	presignClient := s3.NewPresignClient(s3Client)

	return &S3Storage{
		presignClient:  presignClient,
		bucket:         cfg.AWSS3Bucket,
		uploadExpiry:   time.Duration(uploadMins) * time.Minute,
		downloadExpiry: time.Duration(downloadMins) * time.Minute,
	}, nil
}

// GenerateUploadURL returns a pre-signed PUT URL for uploading a CV.
// The key is the S3 object key e.g. "cvs/candidate-id/filename.pdf"
func (s *S3Storage) GenerateUploadURL(ctx context.Context, key string) (string, error) {
	req, err := s.presignClient.PresignPutObject(ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(s.bucket),
			Key:         aws.String(key),
			ContentType: aws.String("application/pdf"),
		},
		s3.WithPresignExpires(s.uploadExpiry),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate upload URL: %w", err)
	}
	return req.URL, nil
}

// GenerateDownloadURL returns a pre-signed GET URL for downloading a CV.
func (s *S3Storage) GenerateDownloadURL(ctx context.Context, key string) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(s.downloadExpiry),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}
	return req.URL, nil
}

// DeleteObject removes a CV from S3 when an application is withdrawn.
func (s *S3Storage) DeleteObject(ctx context.Context, key string) error {
	s3Client := s3.New(s3.Options{})
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}
