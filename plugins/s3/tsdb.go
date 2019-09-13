package s3

import ( 
	"bytes"
	"fmt"
	"io"
	"strings"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/golang/protobuf/proto"
	"github.com/stripe/veneur/samplers"
)

type TSDBEncoder struct{}

func (t *TSDBEncoder) Encode(metrics []samplers.InterMetric, hostName string, interval int) (io.ReadSeeker, error) {
	return EncodeInterMetricsTSDB(metrics, hostName, interval)
}

func createTSDBLabelPairs(tags []string) (label []*dto.LabelPair, err error) {
	for _, keyPair := range tags {
		tag := strings.Split(keyPair, ":")
		if len(tag) != 2 {
			err = fmt.Errorf("Did not produce a key pair of name and value %s", tag)
		}
		labelPair := &dto.LabelPair{
			Name: proto.String(tag[0]),
			Value: proto.String(tag[1]),
		}
		label = append(label, labelPair)
	}
	return label, err
} 

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

func EncodeInterMetricTSDB(d samplers.InterMetric, out *bytes.Buffer, hostName string, interval int) error {
	labelPairs, err := createTSDBLabelPairs(d.Tags)
	if err != nil {
		return err
	}
	var metricType dto.MetricType
	metricValue := d.Value
	metricName := d.Name
	switch d.Type {
	case samplers.CounterMetric:
		metricValue = d.Value / float64(interval)
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


func EncodeInterMetricsTSDB(metrics []samplers.InterMetric, hostname string, interval int) (io.ReadSeeker, error) {
	out := &bytes.Buffer{}
	var err error
	for _, metric := range metrics {
		err = EncodeInterMetricTSDB(metric, out, hostname, interval)
	}
	return bytes.NewReader(out.Bytes()), err
}