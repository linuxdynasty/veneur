package s3

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/stripe/veneur/samplers"
)

type WaveFrontMetricLine struct {
	Name      string
	Value     float64
	TimeStamp int64
	Source    string
	PointTags string
}

// String creates the WaveFront metric in string format.
//<metricName> <metricValue> [<timestamp>] source=<source> [pointTags]
func (w *WaveFrontMetricLine) String() string {
	return fmt.Sprintf("%s %.2f %d source=%s %s\n", w.Name, w.Value, w.TimeStamp, w.Source, w.PointTags)
}

type WaveFrontEncoder struct{}

func (w *WaveFrontEncoder) Encode(metrics []samplers.InterMetric, hostName string, interval int) (io.ReadSeeker, error) {
	return EncodeInterMetricsWaveFront(metrics, hostName, interval)
}

func createPointTags(tags []string) (pointTags string, err error) {
	var sb strings.Builder
	for _, keyPair := range tags {
		tag := strings.Split(keyPair, ":")
		if len(tag) != 2 {
			err = fmt.Errorf("Did not produce a key pair of name and value %s", tag)
		}
		sb.WriteString(fmt.Sprintf("%s=\"%s\" ", tag[0], tag[1]))
	}
	pointTags = strings.TrimSpace(sb.String())
	return pointTags, err
}

func EncodeInterMetricWaveFront(d samplers.InterMetric, out *bytes.Buffer, hostName string, interval int) error {
	tags, err := createPointTags(d.Tags)
	if err != nil {
		return err
	}
	metric := &WaveFrontMetricLine{
		Name:      d.Name,
		Value:     d.Value,
		TimeStamp: d.Timestamp,
		Source:    hostName,
		PointTags: tags,
	}
	if d.Type == samplers.CounterMetric {
		metric.Value = d.Value / float64(interval)
	}
	out.Write([]byte(metric.String()))
	return err
}

func EncodeInterMetricsWaveFront(metrics []samplers.InterMetric, hostName string, interval int) (io.ReadSeeker, error) {
	out := &bytes.Buffer{}
	var err error
	for _, metric := range metrics {
		err = EncodeInterMetricWaveFront(metric, out, hostName, interval)
	}
	return bytes.NewReader(out.Bytes()), err
}
