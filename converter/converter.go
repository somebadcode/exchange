package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

type NoExchangeRateError struct {
	BaseCurrency string
	Currency     string
}

func (e NoExchangeRateError) Error() string {
	return fmt.Sprintf("no exchange rate for %s from %s", e.BaseCurrency, e.Currency)
}

var defaultEndpoint = &url.URL{
	Scheme: "https",
	Host:   "openexchangerates.org",
	Path:   "api/latest.json",
}

type Converter struct {
	Client         *http.Client
	TimeToLive     time.Duration
	Endpoint       *url.URL
	BaseCurrency   string
	AppID          string
	CacheDirectory string
	data           Data
	exchangeRates  map[string]*big.Rat
	lock           sync.RWMutex
}

func (c *Converter) Validate() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.Client == nil {
		jar, err := cookiejar.New(&cookiejar.Options{
			PublicSuffixList: publicsuffix.List,
		})

		if err != nil {
			return fmt.Errorf("creating cookie jar: %w", err)
		}

		c.Client = &http.Client{
			Transport:     http.DefaultTransport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
			Jar:           jar,
			Timeout:       10 * time.Second,
		}
	}

	if c.Endpoint == nil {
		c.Endpoint = defaultEndpoint
	}

	c.BaseCurrency = strings.TrimSpace(strings.ToLower(c.BaseCurrency))

	if c.TimeToLive == 0 {
		c.TimeToLive = 10 * time.Minute
	}

	if c.CacheDirectory == "" {
		p, err := os.UserCacheDir()
		if err != nil {
			return fmt.Errorf("getting cache directory: %w", err)
		}

		c.CacheDirectory = filepath.Join(p, "openexchangerates")
	}

	return nil
}

func (c *Converter) downloadData(ctx context.Context) error {
	u := &url.URL{
		Scheme: c.Endpoint.Scheme,
		Host:   c.Endpoint.Host,
		Path:   c.Endpoint.Path,
	}

	u.Query().Set("base", c.BaseCurrency)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Token "+c.AppID)

	var resp *http.Response

	resp, err = c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&c.data)
	if err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	err = c.updateExchangeRates()
	if err != nil {
		return err
	}

	return nil
}

func (c *Converter) updateExchangeRates() error {
	var err error

	c.exchangeRates = make(map[string]*big.Rat, len(c.data.Rates))

	for k, v := range c.data.Rates {
		c.exchangeRates[k], err = ParseRat(v.String())
		if err != nil {
			return fmt.Errorf("parsing rate: %w", err)
		}
	}

	return nil
}

func (c *Converter) Update(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	cacheFilename := filepath.Join(c.CacheDirectory, c.BaseCurrency+".json")

	if c.data.Timestamp == 0 {
		_ = c.data.Load(cacheFilename)

		err := c.updateExchangeRates()
		if err != nil {
			return err
		}
	}

	if time.Now().Sub(c.data.Time()) < c.TimeToLive {
		return nil
	}

	err := c.downloadData(ctx)
	if err != nil {
		return fmt.Errorf("downloading data: %w", err)
	}

	err = c.data.Save(cacheFilename)
	if err != nil {
		return fmt.Errorf("caching data: %w", err)
	}

	return nil
}

func (c *Converter) ConvertTo(ctx context.Context, currency string, qty *big.Rat) (*big.Rat, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	if err := c.Update(ctx); err != nil {
		return nil, err
	}

	currency = strings.TrimSpace(strings.ToUpper(currency))

	c.lock.RLock()
	defer c.lock.RUnlock()

	rate, ok := c.exchangeRates[currency]
	if !ok {
		return nil, &NoExchangeRateError{
			BaseCurrency: c.BaseCurrency,
			Currency:     currency,
		}
	}

	result := &big.Rat{}

	result.Quo(rate, qty)

	return result, nil
}
