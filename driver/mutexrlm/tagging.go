package mutexrlm

// // SingleTagging is a faster version of [MultiTagging] for situations
// // where only one [Tagger] is sufficient.
// type SingleTagging struct {
// 	failure  error
// 	interval time.Duration
// 	rate     rate.Rate
// 	limit    float64
//
// 	mu sync.Mutex
// 	bucket
// 	taggedBucketMap
// }
//
// // NewSingleTagging initializes a [SingleTagging] rate limiter.
// func NewSingleTagging(withOptions ...Option) (*SingleTagging, error) {
// 	o, err := newOptions(append(
// 		withOptions,
// 		func(o *options) error { // validate
// 			if len(o.Tagging) != 1 {
// 				return errors.New("single-tagged rate limiter must be initiated with exactly one tagger")
// 			}
// 			return nil
// 		},
// 	)...)
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot create single-tagged rate limiter: %w", err)
// 	}
//
// 	s := &SingleTagging{
// 		failure: NewTooManyRequestsError(
// 			fmt.Errorf("rate limiter %q ran out of tokens", o.Supervising.Name)),
// 		rate:            NewRate(o.Supervising.Limit, o.Supervising.Interval),
// 		limit:           o.Supervising.Limit,
// 		interval:        o.Supervising.Interval,
// 		mu:              sync.Mutex{},
// 		bucket:          bucket{},
// 		taggedBucketMap: o.Tagging[0],
// 	}
//
// 	if o.CleanUpContext == nil {
// 		o.CleanUpContext = context.Background()
// 	}
// 	go s.purgeLoop(o.CleanUpContext, o.CleanUpPeriod)
// 	return s, nil
// }
//
// // Rate returns discriminating [rate.Rate] or global [rate.Rate], whichever is slower.
// func (d *SingleTagging) Rate() rate.Rate {
// 	if d.taggedBucketMap.rate < d.rate {
// 		return d.taggedBucketMap.rate
// 	}
// 	return d.rate
// }
//
// // Take first takes from supervising bucket, and then from the tagged bucket map.
// func (d *SingleTagging) Take(r *http.Request) (err error) {
// 	from := time.Now()
// 	d.mu.Lock()
// 	defer d.mu.Unlock()
//
// 	if !d.bucket.Take(
// 		d.limit,
// 		d.rate,
// 		from,
// 		from.Add(d.interval),
// 	) {
// 		err = d.failure
// 	}
// 	return NewTooManyRequestsError(err, d.taggedBucketMap.Take(r, from))
// }
//
// func (d *SingleTagging) purgeLoop(ctx context.Context, interval time.Duration) {
// 	var t time.Time
// 	ticker := time.NewTicker(interval)
// 	defer ticker.Stop()
//
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case t = <-ticker.C:
// 			d.mu.Lock()
// 			d.bucketMap.Purge(t)
// 			d.mu.Unlock()
// 		}
// 	}
// }
//
// // MultiTagging rate limiter takes from the superivising bucket and a list of tagged bucket maps. It is the most flexible of the [RateLimiter]s and the least performant.
// type MultiTagging struct {
// 	failure  error
// 	interval time.Duration
// 	rate     rate.Rate
// 	limit    float64
//
// 	mu sync.Mutex
// 	bucket
// 	taggedBucketMaps []taggedBucketMap
// }
//
// // NewMultiTagging initializes a [MultiTagging] rate limiter.
// func NewMultiTagging(withOptions ...Option) (*MultiTagging, error) {
// 	o, err := newOptions(append(
// 		withOptions,
// 		func(o *options) error { // validate
// 			if len(o.Tagging) < 2 {
// 				return errors.New("tagged rate limiter must be initiated with more than one tagger")
// 			}
// 			return nil
// 		},
// 	)...)
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot create tagged rate limiter: %w", err)
// 	}
//
// 	m := &MultiTagging{
// 		failure: NewTooManyRequestsError(
// 			fmt.Errorf("rate limiter %q ran out of tokens", o.Supervising.Name)),
// 		rate:             NewRate(o.Supervising.Limit, o.Supervising.Interval),
// 		limit:            o.Supervising.Limit,
// 		interval:         o.Supervising.Interval,
// 		mu:               sync.Mutex{},
// 		bucket:           bucket{},
// 		taggedBucketMaps: o.Tagging,
// 	}
//
// 	if o.CleanUpContext == nil {
// 		o.CleanUpContext = context.Background()
// 	}
// 	go m.purgeLoop(o.CleanUpContext, o.CleanUpPeriod)
// 	return m, nil
// }
//
// // Rate returns discriminating [rate.Rate] or global [rate.Rate], whichever is slower.
// func (d *MultiTagging) Rate() (r rate.Rate) {
// 	r = d.rate
// 	for _, child := range d.taggedBucketMaps {
// 		if child.rate < r {
// 			r = child.rate
// 		}
// 	}
// 	return
// }
//
// // Take first takes from supervising bucket, and then from the tagged bucket maps.
// func (d *MultiTagging) Take(r *http.Request) (err error) {
// 	from := time.Now()
// 	d.mu.Lock()
// 	defer d.mu.Unlock()
//
// 	if !d.bucket.Take(
// 		d.limit,
// 		d.rate,
// 		from,
// 		from.Add(d.interval),
// 	) {
// 		err = d.failure
// 	}
//
// 	l := len(d.taggedBucketMaps)
// 	cerr := make([]error, l+1)
// 	cerr[l] = err // last cell is supervising limit error
// 	for i, child := range d.taggedBucketMaps {
// 		cerr[i] = child.Take(r, from)
// 	}
//
// 	return NewTooManyRequestsError(cerr...)
// }
//
// func (d *MultiTagging) purgeLoop(ctx context.Context, interval time.Duration) {
// 	var t time.Time
// 	ticker := time.NewTicker(interval)
// 	defer ticker.Stop()
//
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case t = <-ticker.C:
// 			d.mu.Lock()
// 			for _, child := range d.taggedBucketMaps {
// 				child.bucketMap.Purge(t)
// 			}
// 			d.mu.Unlock()
// 		}
// 	}
// }
