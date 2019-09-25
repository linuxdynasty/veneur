package s3

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"time"

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

type WaveFrontEncoder struct {
	FileNameType      string
	FileNameExtension string
	FileNameStructure string
	Compress          bool
}

func (w *WaveFrontEncoder) Encode(metrics []samplers.InterMetric, hostName string, interval int) (io.ReadSeeker, error) {
	return EncodeInterMetricsWaveFront(metrics, hostName, interval, w.Compress)
}

func (w *WaveFrontEncoder) KeyName(hostname string) (string, error) {
	tNow := time.Now()
	return KeyName(hostname, w.FileNameStructure, w.FileNameType, w.FileNameExtension, w.Compress, tNow)
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

func EncodeInterMetricWaveFront(d samplers.InterMetric, out io.Writer, hostName string, interval int) error {
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

func EncodeInterMetricsWaveFront(metrics []samplers.InterMetric, hostName string, interval int, compress bool) (io.ReadSeeker, error) {
	out := &bytes.Buffer{}
	var err error
	for _, metric := range metrics {
		if compress == true {
			gzw := gzip.NewWriter(out)
			err = EncodeInterMetricWaveFront(metric, gzw, hostName, interval)
			gzw.Flush()
			gzw.Close()
		} else {
			err = EncodeInterMetricWaveFront(metric, out, hostName, interval)
		}
	}
	return bytes.NewReader(out.Bytes()), err
}
