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

// Package discard provides a no-op metrics backend.
package discard

import "github.com/cometbft/cometbft/v2/libs/metrics"

type counter struct{}

// NewCounter returns a new no-op counter.
func NewCounter() metrics.Counter { return counter{} }

// With implements Counter.
func (c counter) With(...string) metrics.Counter { return c }

// Add implements Counter.
func (counter) Add(float64) {}

type gauge struct{}

// NewGauge returns a new no-op gauge.
func NewGauge() metrics.Gauge { return gauge{} }

// With implements Gauge.
func (g gauge) With(...string) metrics.Gauge { return g }

// Set implements Gauge.
func (gauge) Set(float64) {}

// Add implements metrics.Gauge.
func (gauge) Add(float64) {}

type histogram struct{}

// NewHistogram returns a new no-op histogram.
func NewHistogram() metrics.Histogram { return histogram{} }

// With implements Histogram.
func (h histogram) With(...string) metrics.Histogram { return h }

// Observe implements histogram.
func (histogram) Observe(float64) {}
