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

// Package teststat provides helpers for testing metrics backends.
package teststat

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"strings"

	"github.com/cometbft/cometbft/libs/metrics"
)

// TestCounter puts some deltas through the counter, and then calls the value
// func to check that the counter has the correct final value.
func TestCounter(counter metrics.Counter, value func() float64) error {
	want := FillCounter(counter)
	if have := value(); want != have {
		return fmt.Errorf("want %f, have %f", want, have)
	}

	return nil
}

// FillCounter puts some deltas through the counter and returns the total value.
func FillCounter(counter metrics.Counter) float64 {
	a := rand.Perm(100)
	n := rand.Intn(len(a)) //nolint:gosec

	var want float64
	for i := 0; i < n; i++ {
		f := float64(a[i])
		counter.Add(f)
		want += f
	}
	return want
}

// TestGauge puts some values through the gauge, and then calls the value func
// to check that the gauge has the correct final value.
func TestGauge(gauge metrics.Gauge, value func() []float64) error {
	a := rand.Perm(100)
	n := rand.Intn(len(a)) //nolint:gosec

	var want []float64
	for i := 0; i < n; i++ {
		f := float64(a[i])
		gauge.Set(f)
		want = append(want, f)
	}

	for i := 0; i < n; i++ {
		f := float64(a[i])
		gauge.Add(f)
		want = append(want, want[len(want)-1]+f)
	}

	have := value()

	switch len(have) {
	case 0:
		return errors.New("got 0 values")
	case 1: // provider doesn't support multi value
		if have[0] != want[len(want)-1] {
			return fmt.Errorf("want %f, have %f", want, have)
		}
	default: // provider support multi value gauges
		sort.Float64s(want)
		sort.Float64s(have)
		if !reflect.DeepEqual(want, have) {
			return fmt.Errorf("want %f, have %f", want, have)
		}
	}

	return nil
}

// TestHistogram puts some observations through the histogram, and then calls
// the quantiles func to checks that the histogram has computed the correct
// quantiles within some tolerance.
func TestHistogram(histogram metrics.Histogram, quantiles func() (p50, p90, p95, p99 float64), tolerance float64) error {
	PopulateNormalHistogram(histogram, rand.Int()) //nolint:gosec

	want50, want90, want95, want99 := normalQuantiles()
	have50, have90, have95, have99 := quantiles()

	var errs []string
	if want, have := want50, have50; !cmp(want, have, tolerance) {
		errs = append(errs, fmt.Sprintf("p50: want %f, have %f", want, have))
	}
	if want, have := want90, have90; !cmp(want, have, tolerance) {
		errs = append(errs, fmt.Sprintf("p90: want %f, have %f", want, have))
	}
	if have95 > 0 { // prometheus doesn't compute p95
		if want, have := want95, have95; !cmp(want, have, tolerance) {
			errs = append(errs, fmt.Sprintf("p95: want %f, have %f", want, have))
		}
	}
	if want, have := want99, have99; !cmp(want, have, tolerance) {
		errs = append(errs, fmt.Sprintf("p99: want %f, have %f", want, have))
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

var (
	// Count is the number of observations.
	Count = 12345

	// Mean is the center of the normal distribution of observations.
	Mean = 500

	// Stdev of the normal distribution of observations.
	Stdev = 25
)

// ExpectedObservationsLessThan returns the number of observations that should
// have a value less than or equal to the given value, given a normal
// distribution of observations described by Count, Mean, and Stdev.
func ExpectedObservationsLessThan(bucket int64) int64 {
	// https://code.google.com/p/gostat/source/browse/stat/normal.go
	cdf := ((1.0 / 2.0) * (1 + math.Erf((float64(bucket)-float64(Mean))/(float64(Stdev)*math.Sqrt2))))
	return int64(cdf * float64(Count))
}

func cmp(want, have, tol float64) bool {
	return (math.Abs(want-have) / want) <= tol
}
