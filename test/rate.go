package test

import (
	"io"
	"net/http"
	"net/http/httptest"
)

func captureResponse(h http.Handler, r *http.Request) *http.Response {
	// req := httptest.NewRequest(http.MethodGet, "/upper?word=abc", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Result()
}

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "hello world")
})

// func TestLimit(t *testing.T) {
//
// 	cases := []struct {
// 		StatusCode int
// 		Sleep      time.Duration
// 	}{
// 		{StatusCode: http.StatusOK, Sleep: 0},
// 		{StatusCode: http.StatusOK, Sleep: 0},
// 		{StatusCode: http.StatusTooManyRequests, Sleep: time.Millisecond * 1050},
// 		{StatusCode: http.StatusOK, Sleep: 0},
// 		{StatusCode: http.StatusTooManyRequests, Sleep: time.Millisecond * 1050},
// 		{StatusCode: http.StatusOK, Sleep: 0},
// 		{StatusCode: http.StatusTooManyRequests, Sleep: 0},
// 	}
//
// 	handler := Must(NewBasic(
// 		testHandler,
// 		WithLimit(2, time.Second),
// 	))
//
// 	request := httptest.NewRequest(http.MethodGet, "/", nil)
// 	for i, c := range cases {
// 		r := captureResponse(handler, request)
// 		if r.StatusCode != c.StatusCode {
// 			t.Fatalf("rate limiter step %d failed: %d does not match %d",
// 				i+1, r.StatusCode, c.StatusCode)
// 		}
// 		time.Sleep(c.Sleep)
// 	}
// }
