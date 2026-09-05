//go:build !js

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/clientcredentials"
	"github.com/aetherplatform/aether-go/storage"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	tokens, err := clientcredentials.New(clientcredentials.Config{
		TokenURL:     required("AETHER_SDK_SANDBOX_TOKEN_URL"),
		ClientID:     required("AETHER_SDK_SANDBOX_CLIENT_ID"),
		ClientSecret: required("AETHER_SDK_SANDBOX_CLIENT_SECRET"),
		Audience:     "aether-storage",
		Capabilities: []string{
			"storage:objects/*:create",
			"storage:objects/*:delete",
			"storage:objects/*:read",
			"storage:objects/*:share",
		},
	})
	check(err)
	client, err := storage.NewClient(aether.Config{BaseURL: required("AETHER_SDK_SANDBOX_STORAGE_URL"), TokenProvider: tokens})
	check(err)

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	payload := []byte("Aether hosted Go SDK sandbox proof " + runID)
	digest := sha256.Sum256(payload)
	checksum := hex.EncodeToString(digest[:])
	request := storage.CreateUploadIntentRequest{
		NamespaceID:  storage.NamespaceID(required("AETHER_SDK_SANDBOX_NAMESPACE_ID")),
		LogicalKey:   storage.LogicalKey("sdk-proof/" + runID + ".txt"),
		Filename:     runID + ".txt",
		ContentType:  "text/plain",
		ExpectedSize: int64(len(payload)),
		ClientSHA256: &checksum,
	}
	wait := storage.WaitForObjectOptions{PollInterval: time.Second, Timeout: 2 * time.Minute}
	result, err := client.UploadObject(ctx, request, bytes.NewReader(payload), int64(len(payload)), storage.UploadObjectOptions{
		SHA256:              checksum,
		WaitForAvailability: &wait,
	})
	check(err)
	defer func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		_, cleanupErr := client.DeleteObject(cleanupContext, result.Object.ID, aether.WithIdempotencyKey("sdk-proof-delete-"+runID))
		check(cleanupErr)
	}()

	download, err := client.CreateObjectDownloadIntent(ctx, result.Object.ID)
	check(err)
	var downloaded bytes.Buffer
	_, err = storage.DownloadDirect(ctx, download.DownloadURL, &downloaded, nil)
	check(err)
	if !bytes.Equal(downloaded.Bytes(), payload) {
		panic("downloaded payload did not match uploaded payload")
	}
	fmt.Printf("Hosted Go SDK sandbox proof passed for object %s.\n", result.Object.ID)
}

func required(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic(name + " is required")
	}
	return value
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
