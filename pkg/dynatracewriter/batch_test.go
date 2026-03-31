package dynatracewriter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func makeTestMetrics(n int) []dynatraceMetric {
	metrics := make([]dynatraceMetric, n)
	for i := 0; i < n; i++ {
		metrics[i] = dynatraceMetric{
			metricKeyName:    "test.metric",
			metricDimensions: map[string]string{"idx": string(rune('a' + (i % 26)))},
			metricValue:      float64(i),
			metricTimeStamp:  time.Now().UnixMilli(),
		}
	}
	return metrics
}

func TestChunkMetrics_VariousSizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		total     int
		chunkSize int
		wantCount int
	}{
		{"exact_fit", 10, 5, 2},
		{"remainder", 10, 3, 4},
		{"single_chunk", 5, 10, 1},
		{"empty", 0, 5, 0},
		{"one_per_chunk", 5, 1, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := makeTestMetrics(tt.total)
			chunks := chunkMetrics(metrics, tt.chunkSize)
			assert.Equal(t, tt.wantCount, len(chunks))

			// Verify all metrics accounted for
			var totalInChunks int
			for _, c := range chunks {
				totalInChunks += len(c)
			}
			assert.Equal(t, tt.total, totalInChunks)
		})
	}
}

func TestSplitByPayloadSize(t *testing.T) {
	t.Parallel()

	// Create metrics that generate a known payload size
	metrics := makeTestMetrics(100)
	payload := generatePayload(metrics)

	// Split at a size smaller than the full payload
	halfSize := len(payload) / 2
	chunks := splitByPayloadSize(metrics, halfSize)

	assert.Greater(t, len(chunks), 1, "should split into multiple chunks")

	// Each chunk payload must be <= halfSize
	for _, chunk := range chunks {
		p := generatePayload(chunk)
		assert.LessOrEqual(t, len(p), halfSize, "chunk payload exceeds max size")
	}

	// All metrics accounted for
	var total int
	for _, c := range chunks {
		total += len(c)
	}
	assert.Equal(t, 100, total)
}

func TestSplitByPayloadSize_FitsInOne(t *testing.T) {
	t.Parallel()

	metrics := makeTestMetrics(5)
	chunks := splitByPayloadSize(metrics, 1024*1024) // 1 MB
	assert.Equal(t, 1, len(chunks))
	assert.Equal(t, 5, len(chunks[0]))
}

func TestBatchSend_ConcurrentSending(t *testing.T) {
	t.Parallel()

	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	metrics := makeTestMetrics(250)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	headers := map[string]string{
		"Content-Type":  "text/plain; charset=utf-8",
		"Authorization": "Api-Token test",
	}

	results := batchSend(metrics, server.URL, headers, 100, 2, logger)

	// 250 metrics / 100 per batch = 3 batches
	assert.Equal(t, 3, len(results))
	for _, r := range results {
		assert.Nil(t, r.err)
	}
	assert.Equal(t, int64(3), atomic.LoadInt64(&requestCount))
}

func TestBatchSend_Empty(t *testing.T) {
	t.Parallel()

	logger := logrus.New()
	results := batchSend(nil, "http://unused", nil, 100, 2, logger)
	assert.Nil(t, results)
}

func TestBatchSend_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	metrics := makeTestMetrics(10)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	results := batchSend(metrics, server.URL, nil, 5, 1, logger)
	assert.Equal(t, 2, len(results))
	for _, r := range results {
		assert.NotNil(t, r.err)
		assert.True(t, strings.Contains(r.err.Error(), "500"))
	}
}

func TestBatchSend_Backpressure(t *testing.T) {
	t.Parallel()

	var maxConcurrent int64
	var currentConcurrent int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt64(&currentConcurrent, 1)
		// Track max concurrency observed
		for {
			old := atomic.LoadInt64(&maxConcurrent)
			if cur <= old || atomic.CompareAndSwapInt64(&maxConcurrent, old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond) // simulate slow endpoint
		atomic.AddInt64(&currentConcurrent, -1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	metrics := makeTestMetrics(500)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	results := batchSend(metrics, server.URL, nil, 100, 2, logger)

	assert.Equal(t, 5, len(results))
	for _, r := range results {
		assert.Nil(t, r.err)
	}
	// Max concurrency should be capped at 2
	assert.LessOrEqual(t, atomic.LoadInt64(&maxConcurrent), int64(2))
}
