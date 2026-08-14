package storage

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestConfigureS3ClientOptionsUsesConfiguredEndpoint(t *testing.T) {
	options := s3.Options{}
	configureS3ClientOptions(&Config{Endpoint: "https://minio.example.test:9000/"})(&options)

	if !options.UsePathStyle {
		t.Fatal("UsePathStyle = false, want true for S3-compatible endpoints")
	}
	if options.EndpointResolver == nil {
		t.Fatal("EndpointResolver is nil, want configured endpoint resolver")
	}

	endpoint, err := options.EndpointResolver.ResolveEndpoint("us-east-1", options.EndpointOptions)
	if err != nil {
		t.Fatalf("ResolveEndpoint() error = %v", err)
	}
	if endpoint.URL != "https://minio.example.test:9000" {
		t.Fatalf("resolved endpoint URL = %q, want %q", endpoint.URL, "https://minio.example.test:9000")
	}
}
