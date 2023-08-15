/*
Package test provides fixtures for testing [oakratelimiter.RequestLimiter] and [oakratelimiter.RateLimiter] implimentations.
*/
package test

// MiddlewareLoadTest runs a stream of requests while [context.Context] is active against a given rate limiting [Middleware]. Ensures that the failure rate roughly matches the expected failure rate. Use this helper to build and test your own rate limiting middlewares.
// func MiddlewareLoadTest(
// 	ctx context.Context,
// 	m func(oakratelimiter.Handler) oakratelimiter.Handler,
// 	r *rate.Rate,
// 	rf RequestFactory,
// 	expectedRejectionRate float64,
// ) func(t *testing.T) {
// 	return func(t *testing.T) {
// 		handler := m(oakratelimiter.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (err error) {
// 			_, err = io.WriteString(w, "hello world")
// 			return err
// 		}))
//
// 		var err error
// 		requests := make(chan *http.Request, 0)
// 		passed := 0
// 		rejected := 0
//
// 		go func(ctx context.Context, requests chan<- *http.Request) {
// 			// generate requests
// 			oneTokenWindow := time.Nanosecond * time.Duration(1/r.PerNanosecond())
// 			ticker := time.NewTicker(oneTokenWindow)
// 			defer ticker.Stop()
//
// 			for {
// 				select {
// 				case <-ctx.Done():
// 					return
// 				case <-ticker.C:
// 					requests <- rf(ctx)
// 				}
// 			}
// 		}(ctx, requests)
//
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				if passed == 0 {
// 					t.Fatal("no requests succeeded:", passed, "out of", rejected)
// 					return
// 				}
// 				if expectedRejectionRate == 0 && rejected > 0 {
// 					t.Fatalf("%d requests were rejected when 0%% rejection rate was expected", rejected)
// 				}
// 				actualRejectionRate := float64(rejected) / float64(passed+rejected)
// 				if !floatComparator(0.05)(expectedRejectionRate, actualRejectionRate) {
// 					t.Logf("proccessed %d requests, %d passed, %d rejected", passed+rejected, passed, rejected)
// 					t.Fatal(
// 						"expected rejection rate is not close enough to the actual",
// 						expectedRejectionRate,
// 						"vs",
// 						actualRejectionRate,
// 					)
// 				}
// 				return
// 			case request := <-requests:
// 				if request == nil {
// 					t.Error("received a <nil> request")
// 					continue
// 				}
// 				w := httptest.NewRecorder()
// 				err = handler.ServeHyperText(w, request)
// 				if err == nil {
// 					passed++
// 					continue
// 				}
//
// 				var httpError oakratelimiter.Error
// 				if !errors.As(err, &httpError) {
// 					t.Fatal("unexpected error:", err)
// 					return
// 				}
// 				if code := httpError.HyperTextStatusCode(); code != http.StatusTooManyRequests {
// 					t.Fatal("status code mismatch:", code, "vs", http.StatusTooManyRequests)
// 					return
// 				}
// 				rejected++
// 			}
// 		}
// 	}
// 	return nil
// }
