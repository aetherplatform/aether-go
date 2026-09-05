package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aetherplatform/aether-go"
)

type UploadObjectIdempotencyKeys struct {
	Create   string
	Complete string
	Abort    string
}

type UploadObjectOptions struct {
	HTTPClient           *http.Client
	IdempotencyKeys      UploadObjectIdempotencyKeys
	MultipartConcurrency int
	SHA256               string
	WaitForAvailability  *WaitForObjectOptions
}

type UploadWorkflowStage string

const (
	UploadStageCreateIntent UploadWorkflowStage = "create_intent"
	UploadStageTransfer     UploadWorkflowStage = "transfer"
	UploadStageComplete     UploadWorkflowStage = "complete"
)

type UploadWorkflowError struct {
	Stage        UploadWorkflowStage
	UploadID     UploadID
	ObjectID     ObjectID
	Cause        error
	CleanupError error
}

func (err *UploadWorkflowError) Error() string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("storage upload failed during %s", err.Stage)
}

func (err *UploadWorkflowError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

type UploadObjectResult struct {
	Intent *UploadIntent
	Object *Object
}

func (client *Client) UploadObject(ctx context.Context, request CreateUploadIntentRequest, body io.ReaderAt, size int64, options UploadObjectOptions) (*UploadObjectResult, error) {
	if body == nil || size < 0 || request.ExpectedSize != size {
		return nil, &TransferError{Code: "size_mismatch", Operation: TransferUpload}
	}
	keys := options.IdempotencyKeys
	if keys.Create == "" {
		keys.Create = newIdempotencyKey("upload_create")
	}
	if keys.Complete == "" {
		keys.Complete = newIdempotencyKey("upload_complete")
	}
	if keys.Abort == "" {
		keys.Abort = newIdempotencyKey("upload_abort")
	}

	stage := UploadStageCreateIntent
	intent, err := client.CreateUploadIntent(ctx, request, aether.WithIdempotencyKey(keys.Create))
	if err != nil {
		return nil, &UploadWorkflowError{Stage: stage, Cause: err}
	}

	stage = UploadStageTransfer
	var completedParts []CompletedPart
	if intent.UploadStrategy == UploadStrategySingle {
		if intent.UploadURL == nil || intent.Multipart != nil {
			return nil, client.uploadFailure(intent, stage, &TransferError{Code: "invalid_upload_intent", Operation: TransferUpload}, keys.Abort)
		}
		reader := io.NewSectionReader(body, 0, size)
		headers := make(http.Header, len(intent.RequiredHeaders))
		for name, value := range intent.RequiredHeaders {
			headers.Set(name, value)
		}
		if err := UploadDirect(ctx, *intent.UploadURL, reader, size, headers, options.HTTPClient); err != nil {
			return nil, client.uploadFailure(intent, stage, err, keys.Abort)
		}
	} else {
		completedParts, err = UploadMultipart(ctx, *intent, body, size, MultipartUploadOptions{Concurrency: options.MultipartConcurrency, HTTPClient: options.HTTPClient})
		if err != nil {
			return nil, client.uploadFailure(intent, stage, err, keys.Abort)
		}
	}

	stage = UploadStageComplete
	completeRequest := CompleteUploadRequest{}
	sha256 := options.SHA256
	if sha256 == "" && request.ClientSHA256 != nil {
		sha256 = *request.ClientSHA256
	}
	if sha256 != "" {
		completeRequest.SHA256 = &sha256
	}
	if completedParts != nil {
		completeRequest.Parts = &completedParts
	}
	object, err := client.CompleteUpload(ctx, intent.UploadID, completeRequest, aether.WithIdempotencyKey(keys.Complete))
	if err != nil {
		return nil, client.uploadFailure(intent, stage, err, keys.Abort)
	}
	if options.WaitForAvailability != nil {
		object, err = client.WaitForObjectAvailable(ctx, intent.ObjectID, *options.WaitForAvailability)
		if err != nil {
			return nil, err
		}
	}
	return &UploadObjectResult{Intent: intent, Object: object}, nil
}

func (client *Client) uploadFailure(intent *UploadIntent, stage UploadWorkflowStage, cause error, abortKey string) *UploadWorkflowError {
	cleanupContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, cleanupErr := client.AbortUpload(cleanupContext, intent.UploadID, aether.WithIdempotencyKey(abortKey))
	return &UploadWorkflowError{Stage: stage, UploadID: intent.UploadID, ObjectID: intent.ObjectID, Cause: cause, CleanupError: cleanupErr}
}

type WaitForObjectOptions struct {
	PollInterval time.Duration
	Timeout      time.Duration
}

type ObjectAvailabilityError struct {
	Code       string
	ObjectID   ObjectID
	LastStatus ObjectStatus
}

func (err *ObjectAvailabilityError) Error() string {
	if err == nil {
		return ""
	}
	if err.Code == "object_deleted" {
		return "object was deleted before it became available"
	}
	return "timed out waiting for object availability"
}

func (client *Client) WaitForObjectAvailable(ctx context.Context, objectID ObjectID, options WaitForObjectOptions) (*Object, error) {
	pollInterval := options.PollInterval
	if pollInterval == 0 {
		pollInterval = time.Second
	}
	timeout := options.Timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}
	if pollInterval < 0 || timeout < 0 {
		return nil, fmt.Errorf("storage: polling interval and timeout must not be negative")
	}
	deadline := time.Now().Add(timeout)
	for {
		object, err := client.GetObject(ctx, objectID)
		if err != nil {
			return nil, err
		}
		switch object.Status {
		case ObjectStatusAvailable:
			return object, nil
		case ObjectStatusDeleted:
			return nil, &ObjectAvailabilityError{Code: "object_deleted", ObjectID: objectID, LastStatus: object.Status}
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, &ObjectAvailabilityError{Code: "availability_timeout", ObjectID: objectID, LastStatus: object.Status}
		}
		if err := waitContext(ctx, min(pollInterval, remaining)); err != nil {
			return nil, err
		}
	}
}

func (client *Client) ObjectPager(params SearchObjectsParams, supplied ...aether.PagerOptions) *aether.Pager[Object] {
	if len(supplied) > 1 {
		return aether.NewPager[Object](nil, supplied...)
	}
	options := aether.PagerOptions{MaxPages: 1_000}
	if len(supplied) == 1 {
		options = supplied[0]
	}
	if options.InitialCursor == "" && params.Cursor != nil {
		options.InitialCursor = *params.Cursor
	}
	return aether.NewPager(func(ctx context.Context, cursor string) (aether.CursorPage[Object], error) {
		request := params
		if cursor == "" {
			request.Cursor = nil
		} else {
			request.Cursor = &cursor
		}
		response, err := client.SearchObjects(ctx, request)
		if err != nil {
			return aether.CursorPage[Object]{}, err
		}
		next := ""
		if response.NextCursor != nil {
			next = *response.NextCursor
		}
		return aether.CursorPage[Object]{Items: response.Objects, NextCursor: next}, nil
	}, options)
}

func newIdempotencyKey(prefix string) string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(value[:])
}

func waitContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
