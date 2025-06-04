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

package teststat

import (
	"math"
	"math/rand"

	"github.com/cometbft/cometbft/v2/libs/metrics"
)

// PopulateNormalHistogram makes a series of normal random observations into the
// histogram. The number of observations is determined by Count. The randomness
// is determined by Mean, Stdev, and the seed parameter.
//
// This is a low-level function, exported only for metrics that don't perform
// dynamic quantile computation, like a Prometheus Histogram (c.f. Summary). In
// most cases, you don't need to use this function, and can use TestHistogram
// instead.
func PopulateNormalHistogram(h metrics.Histogram, seed int) {
	r := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
	for i := 0; i < Count; i++ {
		sample := r.NormFloat64()*float64(Stdev) + float64(Mean)
		if sample < 0 {
			sample = 0
		}
		h.Observe(sample)
	}
}

func normalQuantiles() (p50, p90, p95, p99 float64) {
	return nvq(50), nvq(90), nvq(95), nvq(99)
}

func nvq(quantile int) float64 {
	// https://en.wikipedia.org/wiki/Normal_distribution#Quantile_function
	return float64(Mean) + float64(Stdev)*math.Sqrt2*erfinv(2*(float64(quantile)/100)-1)
}

func erfinv(y float64) float64 {
	// https://stackoverflow.com/questions/5971830/need-code-for-inverse-error-function
	if y < -1.0 || y > 1.0 {
		panic("invalid input")
	}

	var (
		a = [4]float64{0.886226899, -1.645349621, 0.914624893, -0.140543331}
		b = [4]float64{-2.118377725, 1.442710462, -0.329097515, 0.012229801}
		c = [4]float64{-1.970840454, -1.624906493, 3.429567803, 1.641345311}
		d = [2]float64{3.543889200, 1.637067800}
	)

	const y0 = 0.7
	var x, z float64

	switch {
	case math.Abs(y) == 1.0:
		x = -y * math.Log(0.0)
	case y < -y0:
		z = math.Sqrt(-math.Log((1.0 + y) / 2.0))
		x = -(((c[3]*z+c[2])*z+c[1])*z + c[0]) / ((d[1]*z+d[0])*z + 1.0)
	default:
		if y < y0 {
			z = y * y
			x = y * (((a[3]*z+a[2])*z+a[1])*z + a[0]) / ((((b[3]*z+b[3])*z+b[1])*z+b[0])*z + 1.0)
		} else {
			z = math.Sqrt(-math.Log((1.0 - y) / 2.0))
			x = (((c[3]*z+c[2])*z+c[1])*z + c[0]) / ((d[1]*z+d[0])*z + 1.0)
		}
		x -= (math.Erf(x) - y) / (2.0 / math.SqrtPi * math.Exp(-x*x))
		x -= (math.Erf(x) - y) / (2.0 / math.SqrtPi * math.Exp(-x*x))
	}

	return x
}
