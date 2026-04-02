package dynatracewriter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// defaultBatchTimeout is the per-batch HTTP request timeout.
const defaultBatchTimeout = 30 * time.Second

// chunkMetrics splits a slice of dynatraceMetric into chunks of at most chunkSize.
func chunkMetrics(metrics []dynatraceMetric, chunkSize int) [][]dynatraceMetric {
	if chunkSize <= 0 {
		chunkSize = defaultBatchSize
	}
	var chunks [][]dynatraceMetric
	for i := 0; i < len(metrics); i += chunkSize {
		end := i + chunkSize
		if end > len(metrics) {
			end = len(metrics)
		}
		chunks = append(chunks, metrics[i:end])
	}
	return chunks
}

// splitByPayloadSize further splits a chunk if its serialized payload exceeds maxBytes.
func splitByPayloadSize(metrics []dynatraceMetric, maxBytes int) [][]dynatraceMetric {
	if len(metrics) == 0 {
		return nil
	}

	payload := generatePayload(metrics)
	if len(payload) <= maxBytes {
		return [][]dynatraceMetric{metrics}
	}

	// Split in half and recurse
	mid := len(metrics) / 2
	left := splitByPayloadSize(metrics[:mid], maxBytes)
	right := splitByPayloadSize(metrics[mid:], maxBytes)
	return append(left, right...)
}

// batchResult holds the outcome of a single batch send.
type batchResult struct {
	batchIndex int
	count      int
	err        error
	status     string
}

// batchSend splits dynatraceMetrics into batches, enforces payload size limits,
// and sends them concurrently up to maxConcurrency.
func batchSend(
	metrics []dynatraceMetric,
	url string,
	headers map[string]string,
	batchSize int,
	maxConcurrency int,
	client *http.Client,
	logger logrus.FieldLogger,
) []batchResult {
	if len(metrics) == 0 {
		return nil
	}

	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	if maxConcurrency <= 0 {
		maxConcurrency = defaultMaxConcurrentExports
	}

	// Step 1: chunk by count
	chunks := chunkMetrics(metrics, batchSize)

	// Step 2: enforce payload size limit on each chunk
	var finalChunks [][]dynatraceMetric
	for _, chunk := range chunks {
		sized := splitByPayloadSize(chunk, defaultMaxPayloadBytes)
		finalChunks = append(finalChunks, sized...)
	}

	logger.WithField("total_metrics", len(metrics)).
		WithField("batches", len(finalChunks)).
		Debug("Dynatrace: sending metrics in batches")

	// Step 3: send concurrently with semaphore
	sem := make(chan struct{}, maxConcurrency)
	results := make([]batchResult, len(finalChunks))
	var wg sync.WaitGroup

	for i, chunk := range finalChunks {
		wg.Add(1)
		go func(idx int, batch []dynatraceMetric) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			results[idx] = sendBatch(idx, batch, url, headers, client, logger)
		}(i, chunk)
	}

	wg.Wait()
	return results
}

func sendBatch(
	batchIndex int,
	batch []dynatraceMetric,
	url string,
	headers map[string]string,
	client *http.Client,
	logger logrus.FieldLogger,
) batchResult {
	payload := generatePayload(batch)

	ctx, cancel := context.WithTimeout(context.Background(), defaultBatchTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return batchResult{batchIndex: batchIndex, count: len(batch), err: err}
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return batchResult{batchIndex: batchIndex, count: len(batch), err: err}
	}
	defer response.Body.Close()
	io.ReadAll(response.Body) //nolint:errcheck // drain body for connection reuse

	if response.StatusCode >= 400 {
		return batchResult{
			batchIndex: batchIndex,
			count:      len(batch),
			err:        fmt.Errorf("HTTP %d from Dynatrace ingest", response.StatusCode),
			status:     response.Status,
		}
	}

	logger.WithField("batch", batchIndex).
		WithField("metrics", len(batch)).
		WithField("payload_bytes", len(payload)).
		Debug("Dynatrace: batch sent successfully")

	return batchResult{batchIndex: batchIndex, count: len(batch), status: response.Status}
}
