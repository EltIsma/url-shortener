package metrics

import "github.com/prometheus/client_golang/prometheus"


type PrometheusMetrics struct {
	UrlsTotal prometheus.Gauge
	Redirects *prometheus.CounterVec
	RedirectsTotal prometheus.Counter
    SuccessRequest prometheus.Counter
    Info     *prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *PrometheusMetrics {
    m := &PrometheusMetrics{
        UrlsTotal: prometheus.NewGauge(prometheus.GaugeOpts{
            Namespace: "url_shortener",
            Name:      "number_of_shorten_links",
            Help:      "Total number of currently deployed links",
        }),
		Redirects: prometheus.NewCounterVec(prometheus.CounterOpts{
            Namespace: "url_shortener",
            Name:      "redirect_to_original_url",
            Help:      "Number of redirects.",
        }, []string{"original_url"}),
		RedirectsTotal: prometheus.NewCounter(prometheus.CounterOpts{
            Namespace: "url_shortener",
            Name:      "number_of_redirects",
            Help:      "Number of all redirects.",
        }),
        SuccessRequest: prometheus.NewCounter(prometheus.CounterOpts{
            Namespace: "url_shortener",
            Name:      "number_of_success_requests",
            Help:      "Number of success requests",
        }),
		Info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Namespace: "url_shortener",
            Name:      "info",
            Help:      "Information about the My App environment.",
        }, []string{"version"}),
    }
    reg.MustRegister(m.UrlsTotal, m.Redirects, m.Info, m.RedirectsTotal, m.SuccessRequest)
    return m
}