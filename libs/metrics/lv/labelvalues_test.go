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

package lv

import (
	"strings"
	"testing"
)

func TestWith(t *testing.T) {
	var a LabelValues
	b := a.With("a", "1")
	c := a.With("b", "2", "c", "3")

	if want, have := "", strings.Join(a, ""); want != have {
		t.Errorf("With appears to mutate the original LabelValues: want %q, have %q", want, have)
	}
	if want, have := "a1", strings.Join(b, ""); want != have {
		t.Errorf("With does not appear to return the right thing: want %q, have %q", want, have)
	}
	if want, have := "b2c3", strings.Join(c, ""); want != have {
		t.Errorf("With does not appear to return the right thing: want %q, have %q", want, have)
	}
}
