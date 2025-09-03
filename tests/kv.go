package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Environment variables used by this test program:
// - CF_API_TOKEN: Cloudflare API Token with Workers KV Storage Write/Read permissions
// - CF_ACCOUNT_ID: Cloudflare Account ID
// - CF_KV_NAMESPACE_ID: Workers KV Namespace ID
// - CF_KV_TEST_KEY: (optional) key to use for test, default: "golang-rest-api-template:test"
// - CF_KV_TEST_VALUE: (optional) value to write, default: "hello-from-go"
// - CF_KV_EXPIRATION_TTL: (optional) seconds to live (>=60). If set, key will expire after TTL
// - CF_KV_EXPIRATION: (optional) absolute expiration time in seconds since epoch (ignored if TTL is set)
// - CF_KV_BASE_URL: (optional) override API base URL, default: https://api.cloudflare.com/client/v4

type KVClient struct {
	HTTP        *http.Client
	BaseURL     string
	AccountID   string
	NamespaceID string
	Token       string
}

type apiEnvelope struct {
	Success  bool            `json:"success"`
	Errors   []apiError      `json:"errors"`
	Messages []string        `json:"messages"`
	Result   json.RawMessage `json:"result"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *KVClient) Put(ctx context.Context, key string, value []byte, expirationTTL, expiration *int) error {
	endpoint := fmt.Sprintf("%s/accounts/%s/storage/kv/namespaces/%s/values/%s", c.BaseURL, c.AccountID, c.NamespaceID, url.PathEscape(key))
	q := url.Values{}
	if expirationTTL != nil {
		q.Set("expiration_ttl", strconv.Itoa(*expirationTTL))
	} else if expiration != nil {
		q.Set("expiration", strconv.Itoa(*expiration))
	}
	if len(q) > 0 {
		endpoint += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(value))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("PUT failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	// Cloudflare returns an envelope JSON for PUT; verify success=true
	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	if !env.Success {
		return fmt.Errorf("Cloudflare API reported failure: %+v", env.Errors)
	}
	return nil
}

func (c *KVClient) Get(ctx context.Context, key string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/accounts/%s/storage/kv/namespaces/%s/values/%s", c.BaseURL, c.AccountID, c.NamespaceID, url.PathEscape(key))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("key not found")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvIntPtr(name string) *int {
	v := os.Getenv(name)
	if v == "" {
		return nil
	}
	iv, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &iv
}

func main() {
	baseURL := getenv("CF_KV_BASE_URL", "https://api.cloudflare.com/client/v4")
	client := &KVClient{
		HTTP:        &http.Client{Timeout: 15 * time.Second},
		BaseURL:     baseURL,
		AccountID:   os.Getenv("CF_ACCOUNT_ID"),
		NamespaceID: os.Getenv("CF_KV_NAMESPACE_ID"),
		Token:       os.Getenv("CF_API_TOKEN"),
	}

	if client.Token == "" || client.AccountID == "" || client.NamespaceID == "" {
		fmt.Println("CF_API_TOKEN, CF_ACCOUNT_ID, and CF_KV_NAMESPACE_ID are required")
		os.Exit(1)
	}

	key := getenv("CF_KV_TEST_KEY", "golang-rest-api-template:test")
	val := []byte(getenv("CF_KV_TEST_VALUE", "hello-from-go"))
	expTTL := getEnvIntPtr("CF_KV_EXPIRATION_TTL")
	exp := getEnvIntPtr("CF_KV_EXPIRATION")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Putting key=%q value=%q...\n", key, string(val))
	if err := client.Put(ctx, key, val, expTTL, exp); err != nil {
		fmt.Println("Put error:", err)
		os.Exit(1)
	}

	// Small delay to avoid any edge consistency issues
	time.Sleep(200 * time.Millisecond)

	fmt.Printf("Getting key=%q...\n", key)
	got, err := client.Get(ctx, key)
	if err != nil {
		fmt.Println("Get error:", err)
		os.Exit(1)
	}
	fmt.Printf("Got value=%q\n", string(got))

	if !bytes.Equal(got, val) {
		fmt.Println("Value mismatch!")
		os.Exit(1)
	}
	fmt.Println("KV write/read succeeded.")
}