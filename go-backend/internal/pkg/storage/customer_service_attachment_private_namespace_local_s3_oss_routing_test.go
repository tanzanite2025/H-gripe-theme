package storage

import (
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/require"
)

func TestCustomerServiceAttachmentNamespacesUsePrivateS3AndOSSBucketsAndNeverPublicURLs(t *testing.T) {
	for _, key := range []string{
		CustomerServiceConversationAttachmentOriginalsPrefix + "/2026/10/15/photo.png",
		CustomerServiceConversationAttachmentDerivativesPrefix + "/4/thumbnail/v1/2026/10/15/photo.png",
	} {
		require.True(t, IsPrivateObjectKey(key))
		require.True(t, isPrivateUploadPrefix(key))
		s3Store := &s3StorageImpl{config: &Config{Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1", BaseURL: "https://public.example.test"}}
		bucket, err := s3Store.requireBucketForKey(key)
		require.NoError(t, err)
		require.Equal(t, "private-media", bucket)
		require.NotContains(t, s3Store.GetURL(key), "public.example.test")
		s3Store.config.PrivateBucket = ""
		_, err = s3Store.requireBucketForKey(key)
		require.Error(t, err)
		s3Store.config.PrivateBucket = s3Store.config.Bucket
		_, err = s3Store.requireBucketForKey(key)
		require.Error(t, err)
		privateBucket := &oss.Bucket{}
		ossStore := &ossStorageImpl{config: &Config{Bucket: "public-media", PrivateBucket: "private-media", Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", BaseURL: "https://public.example.test"},
			bucket: &oss.Bucket{}, privateBucket: privateBucket}
		require.Same(t, privateBucket, ossStore.bucketForKey(key))
		require.NotContains(t, ossStore.GetURL(key), "public.example.test")
		ossStore.privateBucket = nil
		require.Nil(t, ossStore.bucketForKey(key))
		require.Empty(t, ossStore.GetURL(key))
	}
	for _, prefix := range []string{CustomerServiceConversationAttachmentOriginalsPrefix, CustomerServiceConversationAttachmentDerivativesPrefix} {
		require.True(t, isPrivateUploadPrefix(prefix))
		require.False(t, IsPrivateObjectKey(prefix+"-public/photo.png"))
		require.False(t, IsPrivateObjectKey(prefix+"/../public/photo.png"))
	}
}
