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
	"github.com/aetherplatform/aether-go/events"
	"github.com/aetherplatform/aether-go/identity"
	"github.com/aetherplatform/aether-go/notifications"
	"github.com/aetherplatform/aether-go/storage"
	"github.com/aetherplatform/aether-go/webhooks"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	checkIdentity(ctx)
	checkEvents(ctx)
	checkNotifications(ctx)
	checkWebhooks(ctx)
	checkStorage(ctx)
	fmt.Println("Hosted Go SDK sandbox proof passed for Identity, Events, Notifications, Webhooks, and Storage.")
}

func checkIdentity(ctx context.Context) {
	client, err := identity.NewClient(identity.Config{BaseURL: required("AETHER_SDK_SANDBOX_IDENTITY_URL")})
	check(err)
	configuration, err := client.GetOpenIDConfiguration(ctx)
	check(err)
	if configuration.Issuer == "" || len(configuration.IDTokenSigningAlgorithmsSupported) == 0 || configuration.IDTokenSigningAlgorithmsSupported[0] != "EdDSA" || !contains(configuration.IDTokenSigningAlgorithmsSupported, "RS256") {
		panic("Identity discovery does not advertise the required EdDSA-first compatibility profile")
	}
	keys, err := client.GetOAuthJWKS(ctx)
	check(err)
	if len(keys.Keys) == 0 {
		panic("Identity JWKS is empty")
	}
}

func checkEvents(ctx context.Context) {
	provider := tokenProvider(required("AETHER_SDK_SANDBOX_EVENTS_AUDIENCE"), []string{"events:catalog/*:read"})
	client, err := events.NewClient(aether.Config{BaseURL: required("AETHER_SDK_SANDBOX_EVENTS_URL"), TokenProvider: provider})
	check(err)
	_, err = client.ListEventTypes(ctx)
	check(err)
}

func checkNotifications(ctx context.Context) {
	provider := tokenProvider(required("AETHER_SDK_SANDBOX_NOTIFICATIONS_AUDIENCE"), []string{"notifications:templates/*:read"})
	client, err := notifications.NewClient(aether.Config{BaseURL: required("AETHER_SDK_SANDBOX_NOTIFICATIONS_URL"), TokenProvider: provider})
	check(err)
	_, err = client.ListNotificationTemplates(ctx)
	check(err)
}

func checkWebhooks(ctx context.Context) {
	provider := tokenProvider(required("AETHER_SDK_SANDBOX_WEBHOOKS_AUDIENCE"), []string{"webhooks:subscriptions/*:read"})
	client, err := webhooks.NewClient(aether.Config{BaseURL: required("AETHER_SDK_SANDBOX_WEBHOOKS_URL"), TokenProvider: provider})
	check(err)
	_, err = client.ListWebhookSubscriptions(ctx, webhooks.ListWebhookSubscriptionsParams{})
	check(err)
}

func checkStorage(ctx context.Context) {
	provider := tokenProvider(required("AETHER_SDK_SANDBOX_STORAGE_AUDIENCE"), []string{
		"storage:objects/*:create",
		"storage:objects/*:delete",
		"storage:objects/*:read",
		"storage:objects/*:share",
	})
	client, err := storage.NewClient(aether.Config{BaseURL: required("AETHER_SDK_SANDBOX_STORAGE_URL"), TokenProvider: provider})
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
	fmt.Printf("Storage hosted proof passed for object %s.\n", result.Object.ID)
}

func tokenProvider(audience string, capabilities []string) *clientcredentials.Provider {
	provider, err := clientcredentials.New(clientcredentials.Config{
		TokenURL:     required("AETHER_SDK_SANDBOX_TOKEN_URL"),
		ClientID:     required("AETHER_SDK_SANDBOX_CLIENT_ID"),
		ClientSecret: required("AETHER_SDK_SANDBOX_CLIENT_SECRET"),
		Audience:     audience,
		Capabilities: capabilities,
	})
	check(err)
	return provider
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
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
