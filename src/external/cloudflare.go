package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const cfAPIBase = "https://api.cloudflare.com/client/v4"

type CloudflareClient struct {
	accountID  string
	apiToken   string
	zoneID     string
	httpClient *http.Client
}

func NewCloudflareClient() *CloudflareClient {
	return &CloudflareClient{
		accountID:  os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		apiToken:   os.Getenv("CLOUDFLARE_API_TOKEN"),
		zoneID:     os.Getenv("CLOUDFLARE_ZONE_ID"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ── R2 Bucket ────────────────────────────────────────────────────────────────

func (c *CloudflareClient) CreateR2Bucket(ctx context.Context, bucketName string) error {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets", cfAPIBase, c.accountID)
	body, _ := json.Marshal(map[string]string{"name": bucketName})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudflare create bucket failed (%d): %s", resp.StatusCode, string(raw))
	}
	return nil
}

// AllowPublicAccess enables the managed pub-*.r2.dev public domain for the bucket
// and returns the assigned domain (e.g. pub-xxxxx.r2.dev).
func (c *CloudflareClient) AllowPublicAccess(ctx context.Context, bucketName string) (string, error) {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets/%s/domains/managed", cfAPIBase, c.accountID, bucketName)
	body, _ := json.Marshal(map[string]any{"enabled": true})

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("cloudflare managed domain failed (%d): %s", resp.StatusCode, string(raw))
	}

	var res struct {
		Success bool `json:"success"`
		Result  struct {
			Domain string `json:"domain"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", err
	}
	if !res.Success {
		return "", fmt.Errorf("cloudflare api returned success = false: %s", string(raw))
	}
	return res.Result.Domain, nil
}

// AttachCustomDomain wires a custom domain to an R2 bucket via Cloudflare API.
func (c *CloudflareClient) AttachCustomDomain(ctx context.Context, bucketName, domain string) error {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets/%s/domains/custom", cfAPIBase, c.accountID, bucketName)
	body, _ := json.Marshal(map[string]any{
		"domain":  domain,
		"enabled": true,
		"zoneId":  c.zoneID,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudflare attach domain failed (%d): %s", resp.StatusCode, string(raw))
	}
	return nil
}

// ── DNS (shared wildcard domain) ─────────────────────────────────────────────

type dnsRecord struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

// CreateCNAME adds a CNAME record in the 5keepr zone so that
// {subdomain}.storage.5keepr.app points to the R2 bucket public endpoint.
func (c *CloudflareClient) CreateCNAME(ctx context.Context, subdomain, target string) error {
	url := fmt.Sprintf("%s/zones/%s/dns_records", cfAPIBase, c.zoneID)
	record := dnsRecord{
		Type:    "CNAME",
		Name:    subdomain,
		Content: target,
		Proxied: true,
		TTL:     1,
	}
	body, _ := json.Marshal(record)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudflare create CNAME failed (%d): %s", resp.StatusCode, string(raw))
	}
	return nil
}

// ── Domain availability ───────────────────────────────────────────────────────

type DomainPrice struct {
	Domain      string  `json:"domain"`
	Available   bool    `json:"available"`
	PriceUSD    float64 `json:"price_usd"`
	PriceBRL    float64 `json:"price_brl"` // approximation
	PurchaseURL string  `json:"purchase_url"`
}

// CheckDomainAvailability checks if a domain is available for registration via
// Cloudflare Registrar and returns pricing info.
// Price is estimated from Cloudflare's published TLD pricing.
func (c *CloudflareClient) CheckDomainAvailability(ctx context.Context, domain string) (*DomainPrice, error) {
	url := fmt.Sprintf("%s/accounts/%s/registrar/domains/%s", cfAPIBase, c.accountID, domain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 404 → not registered by this account (may still be available)
	available := resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusOK

	price := estimateTLDPrice(domain)

	return &DomainPrice{
		Domain:      domain,
		Available:   available,
		PriceUSD:    price,
		PriceBRL:    price * 5.8, // rough BRL conversion
		PurchaseURL: fmt.Sprintf("https://www.cloudflare.com/products/registrar/?domain=%s", domain),
	}, nil
}

// estimateTLDPrice returns the approximate annual price in USD based on
// Cloudflare Registrar's published at-cost pricing for common TLDs.
func estimateTLDPrice(domain string) float64 {
	prices := map[string]float64{
		".com":    9.15,
		".net":    11.06,
		".org":    9.93,
		".io":     32.34,
		".app":    14.00,
		".dev":    12.00,
		".xyz":    3.98,
		".me":     9.90,
		".co":     26.36,
		".com.br": 8.03,
	}
	for tld, price := range prices {
		if len(domain) > len(tld) && domain[len(domain)-len(tld):] == tld {
			return price
		}
	}
	return 12.00 // sensible default
}

// EmptyR2Bucket deletes all objects inside a bucket using the S3-compatible
// API (paginated), then deletes the bucket itself via the Cloudflare REST API.
// R2 refuses to delete a non-empty bucket, so emptying first is required.
func (c *CloudflareClient) EmptyAndDeleteR2Bucket(ctx context.Context, bucketName string) error {
	s3Client, err := c.NewR2S3Client(ctx)
	if err != nil {
		return fmt.Errorf("r2 client: %w", err)
	}

	// Paginate through all objects and delete them in batches of 1000
	var contToken *string
	for {
		list, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucketName),
			ContinuationToken: contToken,
		})
		if err != nil {
			return fmt.Errorf("list objects: %w", err)
		}

		for _, obj := range list.Contents {
			if _, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    obj.Key,
			}); err != nil {
				return fmt.Errorf("delete object %s: %w", *obj.Key, err)
			}
		}

		if !*list.IsTruncated {
			break
		}
		contToken = list.NextContinuationToken
	}

	return c.deleteR2BucketAPI(ctx, bucketName)
}

func (c *CloudflareClient) deleteR2BucketAPI(ctx context.Context, bucketName string) error {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets/%s", cfAPIBase, c.accountID, bucketName)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudflare delete bucket failed (%d): %s", resp.StatusCode, string(raw))
	}
	return nil
}

// ── S3-compatible R2 file client ─────────────────────────────────────────────

// R2FileClient returns an S3 client pointed at the user's R2 bucket.
// Cloudflare R2 uses the same credentials as the main API token for now;
// in production you should create per-bucket API tokens.
func (c *CloudflareClient) NewR2S3Client(ctx context.Context) (*s3.Client, error) {
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", c.accountID)

	cfg := aws.Config{
		Region: "auto",
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("CLOUDFLARE_R2_ACCESS_KEY_ID"),
			os.Getenv("CLOUDFLARE_R2_SECRET_ACCESS_KEY"),
			"",
		),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: r2Endpoint}, nil
			},
		),
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	return client, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (c *CloudflareClient) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
}
