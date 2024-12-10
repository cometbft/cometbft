// The MIT License (MIT)

// Copyright (c) 2015 Peter Bourgon

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package prometheus

import (
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/libs/metrics/teststat"
)

func TestCounter(t *testing.T) {
	s := newServer()
	defer s.Close()

	namespace, subsystem, name := "ns", "ss", "foo"
	re := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `{alpha="alpha-value",beta="beta-value"} ([0-9\.]+)`)

	counter := NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      "This is the help string.",
	}, []string{"alpha", "beta"}).With("beta", "beta-value", "alpha", "alpha-value") // order shouldn't matter

	// minimal delay to allow the prometheus server come up with results and avoid errors during test
	time.Sleep(100 * time.Millisecond)

	value := func() float64 {
		matches := re.FindStringSubmatch(scrape(t, s))
		require.Greater(t, len(matches), 0)
		f, _ := strconv.ParseFloat(matches[1], 64)
		return f
	}

	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	s := newServer()
	defer s.Close()

	namespace, subsystem, name := "aaa", "bbb", "ccc"
	re := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `{foo="bar"} ([0-9\.]+)`)

	gauge := NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      "This is a different help string.",
	}, []string{"foo"}).With("foo", "bar")

	// minimal delay to allow the prometheus server come up with results and avoid errors during test
	time.Sleep(100 * time.Millisecond)

	value := func() []float64 {
		matches := re.FindStringSubmatch(scrape(t, s))
		require.Greater(t, len(matches), 0)
		f, _ := strconv.ParseFloat(matches[1], 64)
		return []float64{f}
	}

	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestSummary(t *testing.T) {
	s := newServer()
	defer s.Close()

	namespace, subsystem, name := "test", "prometheus", "summary"
	re50 := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `{a="a",b="b",quantile="0.5"} ([0-9\.]+)`)
	re90 := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `{a="a",b="b",quantile="0.9"} ([0-9\.]+)`)
	re99 := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `{a="a",b="b",quantile="0.99"} ([0-9\.]+)`)

	summary := NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:  namespace,
		Subsystem:  subsystem,
		Name:       name,
		Help:       "This is the help string for the summary.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"a", "b"}).With("b", "b").With("a", "a")

	// minimal delay to allow the prometheus server come up with results and avoid errors during test
	time.Sleep(100 * time.Millisecond)

	quantiles := func() (float64, float64, float64, float64) {
		buf := scrape(t, s)
		match50 := re50.FindStringSubmatch(buf)
		p50, _ := strconv.ParseFloat(match50[1], 64)
		match90 := re90.FindStringSubmatch(buf)
		p90, _ := strconv.ParseFloat(match90[1], 64)
		match99 := re99.FindStringSubmatch(buf)
		p99, _ := strconv.ParseFloat(match99[1], 64)
		return p50, p90, 0, p99
	}

	if err := teststat.TestHistogram(summary, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	// Prometheus reports histograms as a count of observations that fell into
	// each predefined bucket, with the bucket value representing a global upper
	// limit. That is, the count monotonically increases over the buckets. This
	// requires a different strategy to test.

	s := newServer()
	defer s.Close()

	namespace, subsystem, name := "test", "prometheus", "histogram"
	re := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `_bucket{x="1",le="([0-9]+|\+Inf)"} ([0-9\.]+)`)

	numStdev := 3
	bucketMin := (teststat.Mean - (numStdev * teststat.Stdev))
	bucketMax := (teststat.Mean + (numStdev * teststat.Stdev))
	if bucketMin < 0 {
		bucketMin = 0
	}
	bucketCount := 10
	bucketDelta := (bucketMax - bucketMin) / bucketCount
	buckets := []float64{}
	for i := bucketMin; i <= bucketMax; i += bucketDelta {
		buckets = append(buckets, float64(i))
	}

	histogram := NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      "This is the help string for the histogram.",
		Buckets:   buckets,
	}, []string{"x"}).With("x", "1")

	// minimal delay to allow the prometheus server come up with results and avoid errors during test
	time.Sleep(100 * time.Millisecond)

	// Can't TestHistogram, because Prometheus Histograms don't dynamically
	// compute quantiles. Instead, they fill up buckets. So, let's populate the
	// histogram kind of manually.
	teststat.PopulateNormalHistogram(histogram, rand.Int())

	// Then, we use ExpectedObservationsLessThan to validate.
	for _, line := range strings.Split(scrape(t, s), "\n") {
		match := re.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		bucket, _ := strconv.ParseInt(match[1], 10, 64)
		have, _ := strconv.ParseFloat(match[2], 64)

		want := teststat.ExpectedObservationsLessThan(bucket)
		if match[1] == "+Inf" {
			want = int64(teststat.Count) // special case
		}

		// Unfortunately, we observe experimentally that Prometheus is quite
		// imprecise at the extremes. I'm setting a very high tolerance for now.
		// It would be great to dig in and figure out whether that's a problem
		// with my Expected calculation, or in Prometheus.
		tolerance := 0.5
		if delta := math.Abs(float64(want) - float64(have)); (delta / float64(want)) > tolerance {
			t.Errorf("Bucket %d: want %d, have %d (%.1f%%)", bucket, want, int(have), (100.0 * delta / float64(want)))
		}
	}
}

func TestInconsistentLabelCardinality(t *testing.T) {
	defer func() {
		x := recover()
		if x == nil {
			t.Fatal("expected panic, got none")
		}
		err, ok := x.(error)
		if !ok {
			t.Fatalf("expected error, got %s", reflect.TypeOf(x))
		}
		if want, have := "inconsistent label cardinality", err.Error(); !strings.HasPrefix(have, want) {
			t.Fatalf("want %q, have %q", want, have)
		}
	}()

	NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "test",
		Subsystem: "inconsistent_label_cardinality",
		Name:      "foobar",
		Help:      "This is the help string for the metric.",
	}, []string{"a", "b"}).With(
		"a", "1", "b", "2", "c", "KABOOM!",
	).Add(123)
}

func newServer() *httptest.Server {
	return httptest.NewServer(promhttp.HandlerFor(stdprometheus.DefaultGatherer, promhttp.HandlerOpts{}))
}

func scrape(t *testing.T, s *httptest.Server) string {
	t.Helper()

	resp, err := http.Get(s.URL)
	require.NoError(t, err)
	buf, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = resp.Body.Close()
	require.NoError(t, err)
	return string(buf)
}
