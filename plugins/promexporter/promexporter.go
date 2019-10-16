package promexporter


import {
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stripe/veneur/plugins"
	"github.com/stripe/veneur/samplers"
}

var _ plugins.Plugin =  &PromExporterPlugin

type PromExporterPlugin struct {
	Logger   *logrus.Logger
	EndPointUri string
	Hostname string
	Interval int
	Port int
}

func (p *PromExporterPlugin) Flush(ctx context.Context, metrics []samplers.InterMetric) error {
	http.Handle(p.EndPointUri, promhttp.Handler())
    http.ListenAndServe(":2112", nil)
}