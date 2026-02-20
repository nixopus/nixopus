package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type ImageStore struct {
	client *s3.Client
	bucket string
}

func IsConfigured(cfg types.S3Config) bool {
	return cfg.Bucket != "" && cfg.Endpoint != "" && cfg.AccessKey != "" && cfg.SecretKey != ""
}

func NewImageStore(cfg types.S3Config) (*ImageStore, error) {
	if !IsConfigured(cfg) {
		return nil, fmt.Errorf("S3 configuration is incomplete")
	}

	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			scheme := "https"
			if !cfg.UseSSL {
				scheme = "http"
			}
			return aws.Endpoint{
				URL:               scheme + "://" + cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		},
	)

	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
		awsconfig.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &ImageStore{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func ImageS3Key(orgID, appID, deploymentID uuid.UUID) string {
	return fmt.Sprintf("%s/%s/%s.tar.gz", orgID, appID, deploymentID)
}

// UploadImage streams an image tarball to S3 using multipart upload.
// The reader should produce a gzipped docker save output.
func (s *ImageStore) UploadImage(ctx context.Context, key string, reader io.Reader) (int64, error) {
	uploader := manager.NewUploader(s.client, func(u *manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024 // 64 MB parts
		u.Concurrency = 3
	})

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to upload image to S3: %w", err)
	}

	headOutput, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, fmt.Errorf("image uploaded but failed to get size: %w", err)
	}

	size := *headOutput.ContentLength
	return size, nil
}

// DownloadImage returns a reader for the stored image tarball.
func (s *ImageStore) DownloadImage(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download image from S3: %w", err)
	}
	return output.Body, nil
}

// DeleteImage removes an image tarball from S3.
func (s *ImageStore) DeleteImage(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete image from S3: %w", err)
	}
	return nil
}
