package s3

import (
	"context"
	"errors"
	"io"
	"path"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/sirupsen/logrus"

	"github.com/stripe/veneur/plugins"
	"github.com/stripe/veneur/samplers"
)

// TODO set log level

var _ plugins.Plugin = &S3Plugin{}

type S3Plugin struct {
	Logger   *logrus.Logger
	Svc      s3iface.S3API
	S3Bucket string
	Hostname string
	Interval float64
	Encoder
}

type Encoder interface {
	Encode(metrics []samplers.InterMetric, hostname string, interval float64) (io.ReadSeeker, error)
	KeyName(hostname string) (string, error)
}

func (p *S3Plugin) Flush(ctx context.Context, metrics []samplers.InterMetric) error {
	encodedData, err := p.Encoder.Encode(metrics, p.Hostname, p.Interval)
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			logrus.ErrorKey: err,
			"metrics":       len(metrics),
		}).Error("Could not marshal metrics before posting to s3")
		return err
	}

	err = p.S3Post(p.Hostname, encodedData)
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			logrus.ErrorKey: err,
			"metrics":       len(metrics),
		}).Error("Error posting to s3")
		return err
	}

	p.Logger.WithField("metrics", len(metrics)).Debug("Completed flush to s3")
	return nil
}

func (p *S3Plugin) Name() string {
	return "s3"
}

var S3ClientUninitializedError = errors.New("s3 client has not been initialized")

func (p *S3Plugin) S3Post(hostname string, data io.ReadSeeker) error {
	if p.Svc == nil {
		return S3ClientUninitializedError
	}
	keyName, err := p.Encoder.KeyName(hostname)
	if err != nil {
		return err
	}
	params := &s3.PutObjectInput{
		Bucket: aws.String(p.S3Bucket),
		Key:    aws.String(keyName),
		Body:   data,
	}

	_, err = p.Svc.PutObject(params)
	return err
}

func S3Path(hostname string, ft string) *string {
	t := time.Now()
	filename := strconv.FormatInt(t.Unix(), 10) + "." + ft
	return aws.String(path.Join(t.Format("2006/01/02"), hostname, filename))
}
