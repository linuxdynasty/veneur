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
	Interval int
	Encoder  Encoder
}


type Encoder interface {
	Encode(metrics []samplers.InterMetric, hostname string, interval int) (io.ReadSeeker, error)
}

func (p *S3Plugin) Flush(ctx context.Context, metrics []samplers.InterMetric) error {
	csv, err := p.Encoder.Encode(metrics, p.Hostname, p.Interval)
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			logrus.ErrorKey: err,
			"metrics":       len(metrics),
		}).Error("Could not marshal metrics before posting to s3")
		return err
	}

	err = p.S3Post(p.Hostname, csv, tsvGzFt)
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

type filetype string

const (
	jsonFt  filetype = "json"
	csvFt            = "csv"
	tsvFt            = "tsv"
	tsvGzFt          = "tsv.gz"
)

// S3Bucket name of S3 bucket to post to
var S3Bucket string

var S3ClientUninitializedError = errors.New("s3 client has not been initialized")

func (p *S3Plugin) S3Post(hostname string, data io.ReadSeeker, ft filetype) error {
	if p.Svc == nil {
		return S3ClientUninitializedError
	}
	params := &s3.PutObjectInput{
		Bucket: aws.String(S3Bucket),
		Key:    S3Path(hostname, ft),
		Body:   data,
	}

	_, err := p.Svc.PutObject(params)
	return err
}

func S3Path(hostname string, ft filetype) *string {
	t := time.Now()
	filename := strconv.FormatInt(t.Unix(), 10) + "." + string(ft)
	return aws.String(path.Join(t.Format("2006/01/02"), hostname, filename))
}