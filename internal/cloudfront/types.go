package cloudfront

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
)

type Client struct {
	cfApi   cloudfrontiface.CloudFrontAPI
	timeout time.Duration
}

type CreateInvalidationOutput struct {
	InvalidationId string
	Status         string
}

type GetInvalidationOutput struct {
	CreateTime time.Time
	Status     string
	Paths      []string
}
