package metrics

import (
	"fmt"
	"sync/atomic"
)

var requestsTotal uint64
var cacheHitsTotal uint64
var cacheMissesTotal uint64
var errorsTotal uint64

var responseTimeNanosecondsTotal uint64
var cacheHitResponseTimeNanosecondsTotal uint64
var cacheMissResponseTimeNanosecondsTotal uint64

func IncRequests() {
	atomic.AddUint64(&requestsTotal, 1)
}

func IncCacheHits() {
	atomic.AddUint64(&cacheHitsTotal, 1)
}

func IncCacheMisses() {
	atomic.AddUint64(&cacheMissesTotal, 1)
}

func IncErrors() {
	atomic.AddUint64(&errorsTotal, 1)
}

func AddResponseTimeNanoseconds(ns int64) {
	if ns <= 0 {
		return
	}

	atomic.AddUint64(&responseTimeNanosecondsTotal, uint64(ns))
}

func AddCacheHitResponseTimeNanoseconds(ns int64) {
	if ns <= 0 {
		return
	}

	atomic.AddUint64(&cacheHitResponseTimeNanosecondsTotal, uint64(ns))
}

func AddCacheMissResponseTimeNanoseconds(ns int64) {
	if ns <= 0 {
		return
	}

	atomic.AddUint64(&cacheMissResponseTimeNanosecondsTotal, uint64(ns))
}

func Render() string {
	requests := atomic.LoadUint64(&requestsTotal)
	hits := atomic.LoadUint64(&cacheHitsTotal)
	misses := atomic.LoadUint64(&cacheMissesTotal)
	errors := atomic.LoadUint64(&errorsTotal)

	responseTimeNs := atomic.LoadUint64(&responseTimeNanosecondsTotal)
	hitResponseTimeNs := atomic.LoadUint64(&cacheHitResponseTimeNanosecondsTotal)
	missResponseTimeNs := atomic.LoadUint64(&cacheMissResponseTimeNanosecondsTotal)

	var hitRate float64
	if hits+misses > 0 {
		hitRate = float64(hits) / float64(hits+misses)
	}

	responseTimeSecondsTotal := float64(responseTimeNs) / 1_000_000_000

	var responseTimeSecondsAvg float64
	if requests > 0 {
		responseTimeSecondsAvg = responseTimeSecondsTotal / float64(requests)
	}

	hitResponseTimeSecondsTotal := float64(hitResponseTimeNs) / 1_000_000_000
	missResponseTimeSecondsTotal := float64(missResponseTimeNs) / 1_000_000_000

	var hitResponseTimeSecondsAvg float64
	if hits > 0 {
		hitResponseTimeSecondsAvg = hitResponseTimeSecondsTotal / float64(hits)
	}

	var missResponseTimeSecondsAvg float64
	if misses > 0 {
		missResponseTimeSecondsAvg = missResponseTimeSecondsTotal / float64(misses)
	}

	return fmt.Sprintf(
		`image_proxy_requests_total %d
		image_proxy_cache_hits_total %d
		image_proxy_cache_misses_total %d
		image_proxy_cache_hit_rate %.4f
		image_proxy_errors_total %d
		image_proxy_response_time_seconds_total %.6f
		image_proxy_response_time_seconds_avg %.6f
		image_proxy_cache_hit_response_time_seconds_avg %.6f
		image_proxy_cache_miss_response_time_seconds_avg %.6f
		`,
		requests,
		hits,
		misses,
		hitRate,
		errors,
		responseTimeSecondsTotal,
		responseTimeSecondsAvg,
		hitResponseTimeSecondsAvg,
		missResponseTimeSecondsAvg,
	)
}
