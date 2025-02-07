package postalpool

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/moov-io/watchman/pkg/address"
	"github.com/moov-io/watchman/pkg/search"

	"github.com/moov-io/base/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

// Client provides thread-safe access to a pool of libpostal processes,
// distributing requests across them in a round-robin fashion. Each request
// is automatically routed to the next available worker process.
//
// Requests will be queued up on workers and start to block if every worker
// has an active request. libpostal is single threaded.
//
// Client maintains an address pool to the worker processes. It provides a
// simple API that hides the complexity of working with multiple postal processes.
type Client struct {
	conf      Config
	endpoints []string
	next      atomic.Uint32

	httpClient *http.Client
}

func NewClient(conf Config, endpoints []string) *Client {
	return &Client{
		conf:      conf,
		endpoints: endpoints,
		httpClient: &http.Client{
			Timeout: cmp.Or(conf.RequestTimeout, 10*time.Second),
		},
	}
}

func (c *Client) ParseAddress(ctx context.Context, input string) (search.Address, error) {
	ctx, span := telemetry.StartSpan(ctx, "postalpool-parse-address")
	defer span.End()

	if len(c.endpoints) == 0 {
		span.SetAttributes(attribute.String("postalpool.method", "cgo-zero"))
		return address.ParseAddress(input), nil
	}

	// Simple round-robin including self
	idx := int(c.next.Add(1)) % (len(c.endpoints) + 1)

	// If idx equals last position, use local instance
	if idx == len(c.endpoints) {
		span.SetAttributes(attribute.String("postalpool.method", "cgo-self"))
		return address.ParseAddress(input), nil
	}

	endpoint := c.endpoints[idx]
	span.SetAttributes(
		attribute.String("postalpool.endpoint", endpoint),
		attribute.String("postalpool.method", "binary"),
	)

	var addr search.Address

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"/parse?address="+url.QueryEscape(input), nil)
	if err != nil {
		return addr, fmt.Errorf("creating postal-server request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return addr, fmt.Errorf("HTTP parse to postal-server: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	err = json.NewDecoder(resp.Body).Decode(&addr)
	if err != nil {
		return addr, fmt.Errorf("reading postal-server response: %w", err)
	}
	return addr, nil
}

func (c *Client) healthcheck(ctx context.Context) error {
	ctx, span := telemetry.StartSpan(ctx, "postalpool-client-healthcheck")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, c.conf.StartupTimeout)
	defer cancel()

	results := make(chan error, len(c.endpoints))

	// Try each endpoint every 250ms until success
	for _, endpoint := range c.endpoints {
		go func(ep string) {
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					results <- fmt.Errorf("healthcheck timed out for %s", ep)
					return
				case <-ticker.C:
					_, err := c.ParseAddress(ctx, "")
					if err == nil {
						results <- nil
						return
					}
				}
			}
		}(endpoint)
	}

	// Wait for first success or all failures
	var lastErr error
	for i := 0; i < len(c.endpoints); i++ {
		if err := <-results; err == nil {
			return nil // Found a working endpoint
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("all endpoints failed healthcheck: %w", lastErr)
}
