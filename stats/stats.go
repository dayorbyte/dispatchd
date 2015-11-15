package stats

import (
	"github.com/rcrowley/go-metrics"
	"time"
)

func MakeHistogram(name string) metrics.Histogram {
	return metrics.GetOrRegisterHistogram(
		name,
		metrics.DefaultRegistry,
		metrics.NewUniformSample(10000),
	)
}

func RecordHisto(histo metrics.Histogram, start int64) {
	histo.Update(time.Now().UnixNano() - start)
}

func Start() int64 {
	return time.Now().UnixNano()
}

type Histogram metrics.Histogram
