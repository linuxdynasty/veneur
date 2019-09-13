package s3

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/veneur/samplers"
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
