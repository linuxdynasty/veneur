package s3

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/veneur/samplers"
	. "github.com/stripe/veneur/testhelpers"
)

type WaveFrontTestCase struct {
	Name        string
	InterMetric samplers.InterMetric
	Row         io.Reader
}

func WaveFrontTestCases() []WaveFrontTestCase {

	return []WaveFrontTestCase{
		{
			Name: "BasicGaugeMetric",
			InterMetric: samplers.InterMetric{
				Name:      "a.b.c.max",
				Timestamp: 1476119058,
				Value:     float64(100),
				Tags: []string{"foo:bar",
					"baz:quz"},
				Type: samplers.GaugeMetric,
			},
			Row: strings.NewReader(fmt.Sprintf("a.b.c.max 100.00 1476119058 source=testbox-c3eac9 foo=\"bar\" baz=\"quz\"\n")),
		},
		{
			Name: "BasicGaugeMetricWithCrazyTagValues",
			InterMetric: samplers.InterMetric{
				Name:      "a.b.c.max",
				Timestamp: 1476119058,
				Value:     float64(100),
				Tags: []string{"foo:bar",
					"baz:quz", "msg: CounterMetric: event.http_status_non200s([fake_url])"},
				Type: samplers.GaugeMetric,
			},
			Row: strings.NewReader(fmt.Sprintf("a.b.c.max 100.00 1476119058 source=testbox-c3eac9 foo=\"bar\" baz=\"quz\" msg=\" CounterMetric: event.http_status_non200s([fake_url])\"\n")),
		},
		{
			Name: "BasicCounterMetric",
			InterMetric: samplers.InterMetric{
				Name:      "a.b.c.max",
				Timestamp: 1476119058,
				Value:     float64(100),
				Tags: []string{"foo:bar",
					"baz:quz"},
				Type: samplers.CounterMetric,
			},
			Row: strings.NewReader(fmt.Sprintf("a.b.c.max 10.00 1476119058 source=testbox-c3eac9 foo=\"bar\" baz=\"quz\"\n")),
		},
		{
			// Test that we are able to handle tags which have tab characters in them
			// by quoting the entire field
			// (tags shouldn't do this, but we should handle them properly anyway)
			Name: "TabTag",
			InterMetric: samplers.InterMetric{
				Name:      "a.b.c.count",
				Timestamp: 1476119058,
				Value:     float64(100),
				Tags: []string{"foo:b\tar",
					"baz:quz"},
				Type: samplers.CounterMetric,
			},
			Row: strings.NewReader(fmt.Sprintf("a.b.c.count 10.00 1476119058 source=testbox-c3eac9 foo=\"b	ar\" baz=\"quz\"\n")),
		},
	}
}

func TestEncodeWaveFront(t *testing.T) {
	testCases := WaveFrontTestCases()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			b := &bytes.Buffer{}
			err := EncodeInterMetricWaveFront(tc.InterMetric, b, "testbox-c3eac9", 10)
			fmt.Println(b)
			assert.NoError(t, err)

			assertReadersEqual(t, tc.Row, b)
		})
	}
}

func TestEncodeMetricsWaveFrontCompressed(t *testing.T) {
	const VeneurHostname = "testbox-c3eac9"

	testCases := WaveFrontTestCases()

	metrics := make([]samplers.InterMetric, len(testCases))
	for i, tc := range testCases {
		metrics[i] = tc.InterMetric
	}

	out, err := EncodeInterMetricsWaveFront(metrics, VeneurHostname, 10, true)
	assert.NoError(t, err)

	gzr, err := gzip.NewReader(out)
	assert.NoError(t, err)

	var resB bytes.Buffer
	_, err = resB.ReadFrom(gzr)
	assert.NoError(t, err)

	WaveFrontMatchRecords(resB, metrics, testCases, t)
}

func TestEncodeMetricsWaveFrontUnCompressed(t *testing.T) {
	testCases := WaveFrontTestCases()
	const VeneurHostname = "testbox-c3eac9"

	metrics := make([]samplers.InterMetric, len(testCases))
	for i, tc := range testCases {
		metrics[i] = tc.InterMetric
	}

	out, err := EncodeInterMetricsWaveFront(metrics, VeneurHostname, 10, false)
	assert.NoError(t, err)

	var resB bytes.Buffer
	_, err = resB.ReadFrom(out)
	assert.NoError(t, err)

	WaveFrontMatchRecords(resB, metrics, testCases, t)
}

func WaveFrontMatchRecords(resB bytes.Buffer, metrics []samplers.InterMetric, testCases []WaveFrontTestCase, t *testing.T) {
	listRecords := strings.Split(resB.String(), "\n")
	listRecords = listRecords[:len(listRecords)-1]
	for i, rec := range listRecords {
		t.Logf("record #%d: %s", i, rec)
	}
	t.Log(len(listRecords))

	assert.Equal(t, len(metrics), len(listRecords), "Expected %d records and got %d", len(metrics), len(listRecords))
	for i, tc := range testCases {
		record := listRecords[i] + "\n"
		t.Run(tc.Name, func(t *testing.T) {
			AssertReadersEqual(t, testCases[i].Row, strings.NewReader(record))
		})
	}
}
