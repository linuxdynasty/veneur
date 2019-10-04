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

// WaveFrontMetricLine represents a wavefront metric line.
type WaveFrontMetricLine struct {
	Name      string  // Name of the metric.
	Value     float64 // The value of the metric.
	TimeStamp int64   // The timestamp of the metric.
	Source    string  // The source host of the metric.
	PointTags string  // The Tags aka Labels of the metric.
}

// String creates the WaveFront metric in string format.
//<metricName> <metricValue> [<timestamp>] source=<source> [pointTags]
func (w *WaveFrontMetricLine) String() string {
	return fmt.Sprintf("%s %.2f %d source=%s %s\n", w.Name, w.Value, w.TimeStamp, w.Source, w.PointTags)
}

// WaveFrontEncoder represents a wavefront file and its extensions and how it will be stored.
type WaveFrontEncoder struct {
	FileNameType      string // Either uuid or timestamp.
	FileNameExtension string // `.wavefront` or `""`
	FileNameStructure string // `date_time` or `""`.
	Compress          bool   // Compress and add `.gz` extensio
}

// Encode is an interface to EncodeInterMetricsWaveFront.
func (w *WaveFrontEncoder) Encode(metrics []samplers.InterMetric, hostName string, interval float64) (io.ReadSeeker, error) {
	return EncodeInterMetricsWaveFront(metrics, hostName, interval, w.Compress)
}

// KeyName is an interface to KeyName.
func (w *WaveFrontEncoder) KeyName(hostname string) (string, error) {
	tNow := time.Now()
	return KeyName(hostname, w.FileNameStructure, w.FileNameType, w.FileNameExtension, w.Compress, tNow)
}

// createPointTags create tags in this format foo=bar baz=faz.
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

// EncodeInterMetricWaveFront will encode one metric at a time and write it out to the io.writer
// in plain text in the expected wavefront format and terminated with a newline.
// For performance, encodeInterMetricTSDB does not flush after every call; the caller is
// expected to flush at the end of the operation cycle.
func EncodeInterMetricWaveFront(d samplers.InterMetric, out io.Writer, hostName string, interval float64) error {
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
		metric.Value = d.Value / interval
	}
	out.Write([]byte(metric.String()))
	return err
}

func EncodeInterMetricsWaveFront(metrics []samplers.InterMetric, hostName string, interval float64, compress bool) (io.ReadSeeker, error) {
	out := &bytes.Buffer{}
	var err error
	var gzw *gzip.Writer
	if compress == true {
		gzw = gzip.NewWriter(out)
	}
	for _, metric := range metrics {
		if compress == true {
			err = EncodeInterMetricWaveFront(metric, gzw, hostName, interval)
		} else {
			err = EncodeInterMetricWaveFront(metric, out, hostName, interval)
		}
	}
	if compress == true {
		gzw.Flush()
		gzw.Close()
	}
	return bytes.NewReader(out.Bytes()), err
}
