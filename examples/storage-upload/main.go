package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/clientcredentials"
	"github.com/aetherplatform/aether-go/storage"
)

func main() {
	ctx := context.Background()
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
	if err != nil {
		panic(err)
	}
	client, err := storage.NewClient(aether.Config{
		BaseURL:       required("AETHER_SDK_SANDBOX_STORAGE_URL"),
		TokenProvider: tokens,
	})
	if err != nil {
		panic(err)
	}
	payload := []byte("hello from the Aether Go SDK")
	result, err := client.UploadObject(ctx, storage.CreateUploadIntentRequest{
		NamespaceID:  storage.NamespaceID(required("AETHER_SDK_SANDBOX_NAMESPACE_ID")),
		LogicalKey:   "examples/hello.txt",
		Filename:     "hello.txt",
		ContentType:  "text/plain",
		ExpectedSize: int64(len(payload)),
	}, bytes.NewReader(payload), int64(len(payload)), storage.UploadObjectOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("created object %s\n", result.Object.ID)
}

func required(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic(name + " is required")
	}
	return value
}
