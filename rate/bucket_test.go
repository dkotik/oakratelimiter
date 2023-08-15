package rate

import (
	"testing"
	"time"
)

func floatComparator(errorMargin float64) func(a, b float64) bool {
	return func(a, b float64) bool {
		return b > a-errorMargin && b < a+errorMargin
	}
}

func TestLeakyBucket(t *testing.T) {
	limit := float64(9)
	at := time.Now()
	interval := time.Second
	r, err := New(limit, interval)
	if err != nil {
		t.Fatal("cannot initiate rate:", err)
	}

	bucket := NewLeakyBucket(at, r, r.Burst())
	cases := []struct {
		Sleep     time.Duration
		Remaining float64
	}{
		{Sleep: 0, Remaining: 8},
		{Sleep: 0, Remaining: 7},
		{Sleep: 0, Remaining: 6},
		{Sleep: time.Second / 9, Remaining: 6},
		{Sleep: time.Second, Remaining: 8},
	}

	comp := floatComparator(0.1)
	for i, c := range cases {
		time.Sleep(c.Sleep)
		at := time.Now()
		bucket.Refill(at, r, limit)
		bucket.Take(1.0)
		if comp(bucket.tokens, c.Remaining) {
			t.Log(i+1, "remaining values match", bucket.tokens, c.Remaining)
		} else {
			t.Fatal(i+1, "remaining values do not match", bucket.tokens, c.Remaining)
		}
	}
}
