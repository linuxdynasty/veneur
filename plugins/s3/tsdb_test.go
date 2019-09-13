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

type TSDBTestCase struct {
	Name        string
	InterMetric samplers.InterMetric
	Row         io.Reader
}

func TSDBTestCases() []TSDBTestCase {

	return []TSDBTestCase{
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
			Row: strings.NewReader(fmt.Sprintf("# TYPE a.b.c.max gauge\na.b.c.max{foo=\"bar\",baz=\"quz\"} 100\n")),
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
			Row: strings.NewReader(fmt.Sprintf("# TYPE a.b.c.max counter\na.b.c.max{foo=\"bar\",baz=\"quz\"} 10\n")),
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
			Row: strings.NewReader(fmt.Sprintf("# TYPE a.b.c.count counter\na.b.c.count{foo=\"b	ar\",baz=\"quz\"} 10\n")),
		},
	}
}

func TestEncodeTSDB(t *testing.T) {
	testCases := TSDBTestCases()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			b := &bytes.Buffer{}
			err := EncodeInterMetricTSDB(tc.InterMetric, b, 10)
			fmt.Println(b)
			assert.NoError(t, err)

			assertReadersEqual(t, tc.Row, b)
		})
	}
}
