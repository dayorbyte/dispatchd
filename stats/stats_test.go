package stats

import (
	"github.com/rcrowley/go-metrics"
	"testing"
)

func TestStats(t *testing.T) {
	// This is basically for code coverage. These are simple wrappers.
	// I'll add more tests once this start needing to do more.
	var m = MakeHistogram("hello")
	if m != metrics.Get("hello") {
		t.Errorf("Got different histogram from MakeHistogram")
	}

	var start = Start()
	RecordHisto(m, start-20)
	if m.Max() < 19 || m.Max() > start {
		t.Errorf("Bad value in histo %d", m.Max())
	}
}
