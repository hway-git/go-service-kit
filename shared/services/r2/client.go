package r2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

var (
	ErrNotFound = errors.New("对象不存在")
	ErrRemote   = errors.New("对象存储操作失败")
)

type ObjectInfo struct {
	Key          string     `json:"key"`
	Size         int64      `json:"size"`
	ContentType  string     `json:"content_type"`
	ETag         string     `json:"etag"`
	LastModified *time.Time `json:"last_modified,omitempty"`
}

type Object struct {
	Info ObjectInfo
	Body io.ReadCloser
}

type ListResult struct {
	Items      []ObjectInfo `json:"items"`
	NextCursor string       `json:"next_cursor"`
	Truncated  bool         `json:"truncated"`
}

type PresignedURL struct {
	URL       string              `json:"url"`
	Method    string              `json:"method"`
	Headers   map[string][]string `json:"headers"`
	ExpiresTs int64               `json:"expires_ts"`
}

type Storage interface {
	Put(context.Context, string, io.Reader, int64, string) (*ObjectInfo, error)
	Get(context.Context, string) (*Object, error)
	Head(context.Context, string) (*ObjectInfo, error)
	Delete(context.Context, string) error
	List(context.Context, string, string, int32) (*ListResult, error)
	PresignGet(context.Context, string, time.Duration) (*PresignedURL, error)
	PresignPut(context.Context, string, string, time.Duration) (*PresignedURL, error)
	PublicURL(string) (string, error)
}

// Client 固定服务端配置的bucket，调用方不能通过请求参数替换endpoint或凭据。
type Client struct {
	api            *s3.Client
	signer         *s3.PresignClient
	bucket         string
	maxUploadBytes int64
	public         *url.URL
}

var _ Storage = (*Client)(nil)

// New 初始化SDK客户端，不发起网络请求；HTTP客户端可注入模拟transport进行普通测试。
func New(cfg Config, httpClient *http.Client) (*Client, error) {
	cfg.defaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	httpCopy := http.Client{Timeout: 30 * time.Second}
	if httpClient != nil {
		httpCopy = *httpClient
	}
	// 签名请求不跟随重定向，避免授权头被转发到其他地址。
	httpCopy.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	api := s3.NewFromConfig(aws.Config{
		Region: "auto", Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		HTTPClient: &httpCopy, RetryMaxAttempts: 3,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(strings.TrimRight(cfg.Endpoint, "/"))
		options.UsePathStyle = true
	})
	client := &Client{api: api, signer: s3.NewPresignClient(api), bucket: cfg.Bucket, maxUploadBytes: cfg.MaxUploadBytes}
	if cfg.PublicBaseURL != "" {
		client.public, _ = url.Parse(cfg.PublicBaseURL)
	}
	return client, nil
}

// ValidateKey 校验共享资源服务及对象操作使用的对象key。
func ValidateKey(key string) error {
	if key == "" || len(key) > 1024 || !utf8.ValidString(key) || strings.HasPrefix(key, "/") || strings.Contains(key, "\\") || strings.TrimSpace(key) == "" {
		return fmt.Errorf("对象key格式无效")
	}
	for _, part := range strings.Split(key, "/") {
		if part == "." || part == ".." {
			return fmt.Errorf("对象key不能包含目录跳转")
		}
	}
	for _, character := range key {
		if character < 0x20 || character == 0x7f {
			return fmt.Errorf("对象key不能包含控制字符")
		}
	}
	return nil
}

func contentType(value string) (string, error) {
	if value == "" {
		return "application/octet-stream", nil
	}
	parsed, params, err := mime.ParseMediaType(value)
	if err != nil {
		return "", fmt.Errorf("文件Content-Type无效")
	}
	return mime.FormatMediaType(parsed, params), nil
}

func remoteError(ctx context.Context, operation string, err error) error {
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	var apiError smithy.APIError
	if errors.As(err, &apiError) && (apiError.ErrorCode() == "NoSuchKey" || apiError.ErrorCode() == "NotFound") {
		return ErrNotFound
	}
	var responseError *smithyhttp.ResponseError
	if errors.As(err, &responseError) && responseError.HTTPStatusCode() == 404 {
		return ErrNotFound
	}
	// 底层错误可能含授权URL或远端正文，只保留明确的操作名。
	return fmt.Errorf("R2%s失败：%w", operation, ErrRemote)
}

func (c *Client) Put(ctx context.Context, key string, body io.Reader, size int64, mediaType string) (*ObjectInfo, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}
	if body == nil || size < 0 {
		return nil, fmt.Errorf("上传正文和大小无效")
	}
	if size > c.maxUploadBytes {
		return nil, fmt.Errorf("上传大小超过 R2 配置上限")
	}
	mediaType, err := contentType(mediaType)
	if err != nil {
		return nil, err
	}
	result, err := c.api.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key), Body: body, ContentLength: aws.Int64(size), ContentType: aws.String(mediaType)})
	if err != nil {
		return nil, remoteError(ctx, "上传", err)
	}
	return &ObjectInfo{Key: key, Size: size, ContentType: mediaType, ETag: strings.Trim(aws.ToString(result.ETag), "\"")}, nil
}

// Get 返回读取流，调用方须关闭Body。
func (c *Client) Get(ctx context.Context, key string) (*Object, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}
	result, err := c.api.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, remoteError(ctx, "读取", err)
	}
	return &Object{Info: ObjectInfo{Key: key, Size: aws.ToInt64(result.ContentLength), ContentType: aws.ToString(result.ContentType), ETag: strings.Trim(aws.ToString(result.ETag), "\""), LastModified: result.LastModified}, Body: result.Body}, nil
}

func (c *Client) Head(ctx context.Context, key string) (*ObjectInfo, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}
	result, err := c.api.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, remoteError(ctx, "查询", err)
	}
	return &ObjectInfo{Key: key, Size: aws.ToInt64(result.ContentLength), ContentType: aws.ToString(result.ContentType), ETag: strings.Trim(aws.ToString(result.ETag), "\""), LastModified: result.LastModified}, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	if err := ValidateKey(key); err != nil {
		return err
	}
	_, err := c.api.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key)})
	err = remoteError(ctx, "删除", err)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

func (c *Client) List(ctx context.Context, prefix, cursor string, limit int32) (*ListResult, error) {
	if prefix != "" {
		if err := ValidateKey(prefix); err != nil {
			return nil, err
		}
	}
	if limit < 1 || limit > 1000 {
		return nil, fmt.Errorf("列举数量须在1至1000之间")
	}
	input := &s3.ListObjectsV2Input{Bucket: aws.String(c.bucket), Prefix: aws.String(prefix), MaxKeys: aws.Int32(limit)}
	if cursor != "" {
		input.ContinuationToken = aws.String(cursor)
	}
	result, err := c.api.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, remoteError(ctx, "列举", err)
	}
	rows := make([]ObjectInfo, 0, len(result.Contents))
	for _, object := range result.Contents {
		rows = append(rows, ObjectInfo{Key: aws.ToString(object.Key), Size: aws.ToInt64(object.Size), ETag: strings.Trim(aws.ToString(object.ETag), "\""), LastModified: object.LastModified})
	}
	return &ListResult{Items: rows, NextCursor: aws.ToString(result.NextContinuationToken), Truncated: aws.ToBool(result.IsTruncated)}, nil
}

func validateExpiry(expiry time.Duration) error {
	if expiry < time.Second || expiry > 7*24*time.Hour {
		return fmt.Errorf("预签名有效期须在1秒至7天之间")
	}
	return nil
}

func (c *Client) PresignGet(ctx context.Context, key string, expiry time.Duration) (*PresignedURL, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}
	if err := validateExpiry(expiry); err != nil {
		return nil, err
	}
	result, err := c.signer.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key)}, func(options *s3.PresignOptions) { options.Expires = expiry })
	if err != nil {
		return nil, remoteError(ctx, "下载签名", err)
	}
	return &PresignedURL{URL: result.URL, Method: result.Method, Headers: result.SignedHeader, ExpiresTs: time.Now().Unix() + int64(expiry/time.Second)}, nil
}

func (c *Client) PresignPut(ctx context.Context, key, mediaType string, expiry time.Duration) (*PresignedURL, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}
	if err := validateExpiry(expiry); err != nil {
		return nil, err
	}
	mediaType, err := contentType(mediaType)
	if err != nil {
		return nil, err
	}
	result, err := c.signer.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key), ContentType: aws.String(mediaType)}, func(options *s3.PresignOptions) { options.Expires = expiry })
	if err != nil {
		return nil, remoteError(ctx, "上传签名", err)
	}
	return &PresignedURL{URL: result.URL, Method: result.Method, Headers: result.SignedHeader, ExpiresTs: time.Now().Unix() + int64(expiry/time.Second)}, nil
}

func (c *Client) PublicURL(key string) (string, error) {
	if err := ValidateKey(key); err != nil {
		return "", err
	}
	if c.public == nil {
		return "", nil
	}
	address := *c.public
	address.Path = strings.TrimRight(address.Path, "/") + "/" + key
	address.RawPath = ""
	return address.String(), nil
}
