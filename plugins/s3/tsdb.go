package s3

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stripe/veneur/samplers"
)

// TSDBEncoder represents a tsdb file and its extensions and how it will be stored.
type TSDBEncoder struct {
	FileNameType      string // Either uuid or timestamp.
	FileNameExtension string // `.tsdb` or `""`
	FileNameStructure string // `date_time` or `""`.
	Compress          bool   // Compress and add `.gz` extension
}

// Encode is an interface to EncodeInterMetricsTSDB
func (t *TSDBEncoder) Encode(metrics []samplers.InterMetric, _hostName string, interval float64) (io.ReadSeeker, error) {
	return EncodeInterMetricsTSDB(metrics, interval, t.Compress)
}

// KeyName is an interface to KeyName
func (t *TSDBEncoder) KeyName(hostname string) (string, error) {
	tNow := time.Now()
	return KeyName(hostname, t.FileNameStructure, t.FileNameType, t.FileNameExtension, t.Compress, tNow)
}

// createTSDBLabelPairs formats the list of strings that are formatted as json tags into a TSDB Label Pair.
// Example of list of strings that can be passed. ['{"Foo": "Bar"}'].
func createTSDBLabelPairs(tags []string) (label []*dto.LabelPair, err error) {
	for _, keyPair := range tags {
		tag := strings.Split(keyPair, ":")
		if len(tag) != 2 {
			err = fmt.Errorf("Did not produce a key pair of name and value %s", tag)
		}
		labelPair := &dto.LabelPair{
			Name:  proto.String(tag[0]),
			Value: proto.String(tag[1]),
		}
		label = append(label, labelPair)
	}
	return label, err
}

// createTSDBMetric creates the tsdb formatted metric and returns it in *dto.MetricFamily and error.
func createTSDBMetric(metricType dto.MetricType, metricValue float64, metricName string, labelPairs []*dto.LabelPair) (*dto.MetricFamily, error) {
	var err error
	metric := &dto.MetricFamily{
		Name: proto.String(metricName),
		Type: &metricType,
	}
	if metricType == dto.MetricType_COUNTER {
		metric.Metric = []*dto.Metric{
			&dto.Metric{
				Label: labelPairs,
				Counter: &dto.Counter{
					Value: proto.Float64(metricValue),
				},
			},
		}
	} else if metricType == dto.MetricType_GAUGE {
		metric.Metric = []*dto.Metric{
			&dto.Metric{
				Label: labelPairs,
				Gauge: &dto.Gauge{
					Value: proto.Float64(metricValue),
				},
			},
		}
	} else {
		err = fmt.Errorf("Invalid MetricType %s", metricType)
	}
	return metric, err
}

// EncodeInterMetricTSDB will encode one metric at a time and write it out to the io.writer
// in plain text in the expected tsdb format and terminated with a newline.
// For performance, encodeInterMetricTSDB does not flush after every call; the caller is
// expected to flush at the end of the operation cycle.
func EncodeInterMetricTSDB(d samplers.InterMetric, out io.Writer, interval float64) error {
	labelPairs, err := createTSDBLabelPairs(d.Tags)
	if err != nil {
		return err
	}
	var metricType dto.MetricType
	metricValue := d.Value
	metricName := d.Name
	switch d.Type {
	case samplers.CounterMetric:
		metricValue = d.Value / interval
		metricType = dto.MetricType_COUNTER
	case samplers.GaugeMetric:
		metricType = dto.MetricType_GAUGE
	default:
		return fmt.Errorf("Encountered an unknown metric type %s", d.Type.String())
	}
	metric, mErr := createTSDBMetric(metricType, metricValue, metricName, labelPairs)
	if mErr != nil {
		return mErr
	}
	_, err = expfmt.MetricFamilyToText(out, metric)
	if err != nil {
		return err
	}
	return err
}

// EncodeInterMetricsTSDB returns a reader containing either the gzipped or plain text representation of the
// InterMetric data, one row per InterMetric.
// the AWS sdk requires seekable input, so we return a ReadSeeker here.
func EncodeInterMetricsTSDB(metrics []samplers.InterMetric, interval float64, compress bool) (io.ReadSeeker, error) {
	out := &bytes.Buffer{}
	var err error
	for _, metric := range metrics {
		if compress == true {
			gzw := gzip.NewWriter(out)
			err = EncodeInterMetricTSDB(metric, gzw, interval)
			gzw.Flush()
			gzw.Close()
		} else {
			err = EncodeInterMetricTSDB(metric, out, interval)
		}
	}
	return bytes.NewReader(out.Bytes()), err
}
