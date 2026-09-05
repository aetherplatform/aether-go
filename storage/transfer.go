package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type TransferOperation string

const (
	TransferUpload        TransferOperation = "upload"
	TransferMultipartPart TransferOperation = "multipart_part"
	TransferDownload      TransferOperation = "download"
)

type TransferError struct {
	Code       string
	Operation  TransferOperation
	StatusCode int
	PartNumber int64
	Cause      error
}

func (err *TransferError) Error() string {
	if err == nil {
		return ""
	}
	switch err.Code {
	case "aborted":
		return "storage transfer was aborted"
	case "invalid_upload_intent":
		return "storage returned an invalid upload intent"
	case "missing_etag":
		return fmt.Sprintf("multipart part %d did not return an ETag", err.PartNumber)
	case "network_error":
		return "storage transfer failed"
	case "size_mismatch":
		if err.Operation == TransferDownload {
			return "download size does not match the provider response"
		}
		return "upload body size does not match the upload instructions"
	default:
		if err.PartNumber > 0 {
			return fmt.Sprintf("multipart part %d failed with HTTP %d", err.PartNumber, err.StatusCode)
		}
		return fmt.Sprintf("storage transfer failed with HTTP %d", err.StatusCode)
	}
}

func (err *TransferError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

type MultipartUploadOptions struct {
	Concurrency int
	HTTPClient  *http.Client
	Headers     http.Header
}

func UploadDirect(ctx context.Context, uploadURL string, body io.Reader, size int64, headers http.Header, httpClient *http.Client) error {
	if body == nil || size < 0 {
		return &TransferError{Code: "size_mismatch", Operation: TransferUpload}
	}
	request, err := providerRequest(ctx, http.MethodPut, uploadURL, body, size, headers)
	if err != nil {
		return &TransferError{Code: "invalid_upload_intent", Operation: TransferUpload, Cause: err}
	}
	response, err := providerClient(httpClient).Do(request)
	if err != nil {
		return transferRequestError(ctx, TransferUpload, 0, err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &TransferError{Code: "transfer_failed", Operation: TransferUpload, StatusCode: response.StatusCode}
	}
	return nil
}

func DownloadDirect(ctx context.Context, downloadURL string, destination io.Writer, httpClient *http.Client) (int64, error) {
	if destination == nil {
		return 0, &TransferError{Code: "transfer_failed", Operation: TransferDownload, Cause: fmt.Errorf("destination is required")}
	}
	request, err := providerRequest(ctx, http.MethodGet, downloadURL, nil, 0, nil)
	if err != nil {
		return 0, &TransferError{Code: "transfer_failed", Operation: TransferDownload, Cause: err}
	}
	response, err := providerClient(httpClient).Do(request)
	if err != nil {
		return 0, transferRequestError(ctx, TransferDownload, 0, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, response.Body)
		return 0, &TransferError{Code: "transfer_failed", Operation: TransferDownload, StatusCode: response.StatusCode}
	}
	written, err := io.Copy(destination, response.Body)
	if err != nil {
		return written, &TransferError{Code: "network_error", Operation: TransferDownload, StatusCode: response.StatusCode, Cause: err}
	}
	if response.ContentLength >= 0 && written != response.ContentLength {
		return written, &TransferError{Code: "size_mismatch", Operation: TransferDownload, StatusCode: response.StatusCode}
	}
	return written, nil
}

func UploadMultipart(ctx context.Context, intent UploadIntent, body io.ReaderAt, size int64, options MultipartUploadOptions) ([]CompletedPart, error) {
	parts, ranges, err := validateMultipartIntent(intent, body, size)
	if err != nil {
		return nil, err
	}
	concurrency := options.Concurrency
	if concurrency == 0 {
		concurrency = 4
	}
	if concurrency < 1 {
		return nil, fmt.Errorf("storage: multipart concurrency must be positive")
	}
	if concurrency > len(parts) {
		concurrency = len(parts)
	}

	workerContext, cancel := context.WithCancel(ctx)
	defer cancel()
	completed := make([]CompletedPart, len(parts))
	var next int
	var nextMu sync.Mutex
	var firstErr error
	var firstErrOnce sync.Once
	var workers sync.WaitGroup

	worker := func() {
		defer workers.Done()
		for {
			if workerContext.Err() != nil {
				return
			}
			nextMu.Lock()
			index := next
			next++
			nextMu.Unlock()
			if index >= len(parts) {
				return
			}
			part := parts[index]
			byteRange := ranges[index]
			reader := io.NewSectionReader(body, byteRange.start, byteRange.end-byteRange.start)
			request, requestErr := providerRequest(workerContext, http.MethodPut, part.UploadURL, reader, byteRange.end-byteRange.start, mergeHeaders(options.Headers, intent.RequiredHeaders))
			if requestErr != nil {
				requestErr = &TransferError{Code: "invalid_upload_intent", Operation: TransferMultipartPart, PartNumber: part.PartNumber, Cause: requestErr}
			}
			if requestErr == nil {
				var response *http.Response
				response, requestErr = providerClient(options.HTTPClient).Do(request)
				if requestErr == nil {
					_, _ = io.Copy(io.Discard, response.Body)
					response.Body.Close()
					if response.StatusCode < 200 || response.StatusCode >= 300 {
						requestErr = &TransferError{Code: "transfer_failed", Operation: TransferMultipartPart, StatusCode: response.StatusCode, PartNumber: part.PartNumber}
					} else {
						etag := strings.TrimSpace(response.Header.Get("etag"))
						if etag == "" {
							requestErr = &TransferError{Code: "missing_etag", Operation: TransferMultipartPart, StatusCode: response.StatusCode, PartNumber: part.PartNumber}
						} else {
							completed[index] = CompletedPart{PartNumber: part.PartNumber, ETag: etag}
						}
					}
				}
			}
			if requestErr != nil {
				if _, ok := requestErr.(*TransferError); !ok {
					requestErr = transferRequestError(workerContext, TransferMultipartPart, part.PartNumber, requestErr)
				}
				firstErrOnce.Do(func() {
					firstErr = requestErr
					cancel()
				})
				return
			}
		}
	}

	workers.Add(concurrency)
	for range concurrency {
		go worker()
	}
	workers.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, &TransferError{Code: "aborted", Operation: TransferMultipartPart, Cause: err}
	}
	return completed, nil
}

type uploadRange struct {
	start int64
	end   int64
}

func validateMultipartIntent(intent UploadIntent, body io.ReaderAt, size int64) ([]MultipartPart, []uploadRange, error) {
	if intent.UploadStrategy != UploadStrategyMultipart || intent.Multipart == nil || body == nil || size <= 0 || intent.Multipart.PartSize <= 0 {
		return nil, nil, &TransferError{Code: "invalid_upload_intent", Operation: TransferMultipartPart}
	}
	parts := append([]MultipartPart(nil), intent.Multipart.Parts...)
	sort.Slice(parts, func(left, right int) bool { return parts[left].PartNumber < parts[right].PartNumber })
	expectedCount := int((size + intent.Multipart.PartSize - 1) / intent.Multipart.PartSize)
	if len(parts) != expectedCount {
		return nil, nil, &TransferError{Code: "size_mismatch", Operation: TransferMultipartPart}
	}
	ranges := make([]uploadRange, len(parts))
	for index, part := range parts {
		if part.PartNumber != int64(index+1) || strings.TrimSpace(part.UploadURL) == "" {
			return nil, nil, &TransferError{Code: "invalid_upload_intent", Operation: TransferMultipartPart, PartNumber: part.PartNumber}
		}
		start := int64(index) * intent.Multipart.PartSize
		end := min(start+intent.Multipart.PartSize, size)
		ranges[index] = uploadRange{start: start, end: end}
	}
	return parts, ranges, nil
}

func providerRequest(ctx context.Context, method, target string, body io.Reader, contentLength int64, headers http.Header) (*http.Request, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}
	request, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	for name, values := range headers {
		if forbiddenProviderHeader(name) {
			return nil, fmt.Errorf("provider header %s is forbidden", name)
		}
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
	if body != nil {
		request.ContentLength = contentLength
	}
	return request, nil
}

func forbiddenProviderHeader(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "authorization", "proxy-authorization", "x-request-id", "x-correlation-id":
		return true
	default:
		return false
	}
}

func mergeHeaders(additional http.Header, required map[string]string) http.Header {
	merged := make(http.Header, len(additional)+len(required))
	for name, values := range additional {
		merged[name] = append([]string(nil), values...)
	}
	for name, value := range required {
		merged.Set(name, value)
	}
	return merged
}

func providerClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return http.DefaultClient
}

func transferRequestError(ctx context.Context, operation TransferOperation, partNumber int64, cause error) *TransferError {
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return &TransferError{Code: "aborted", Operation: operation, PartNumber: partNumber, Cause: cause}
	}
	return &TransferError{Code: "network_error", Operation: operation, PartNumber: partNumber, Cause: cause}
}
