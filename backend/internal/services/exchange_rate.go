package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ExchangeRateService struct {
	Pool       *pgxpool.Pool
	HTTPClient *http.Client
}

const CacheDuration = 24 * time.Hour

func NewExchangeRateService(pool *pgxpool.Pool) *ExchangeRateService {
	return &ExchangeRateService{
		Pool: pool,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *ExchangeRateService) GetRate(ctx context.Context, baseCurrency, targetCurrency string) (float64, error) {
	if baseCurrency == targetCurrency {
		return 1.0, nil
	}

	cachedRate, isStale, err := s.getCachedRate(ctx, baseCurrency, targetCurrency)
	if err == nil && !isStale {
		return cachedRate, nil
	}

	apiRate, err := s.fetchFromAPI(ctx, baseCurrency, targetCurrency)
	if err != nil {
		if cachedRate > 0 {
			return cachedRate, nil
		}
		return 0, fmt.Errorf("failed to fetch rate and no cache available: %v", err)
	}

	s.updateCache(ctx, baseCurrency, targetCurrency, apiRate)

	return apiRate, nil

}

func (s *ExchangeRateService) getCachedRate(ctx context.Context, base, target string) (float64, bool, error) {
	query := `SELECT rate, fetcched_at 
		FROM exchange_rates_cache
		WHERE base_currency = $1 
			AND target_currency = $2
		LIMIT 1`

	var rate float64
	var fetchedAt time.Time

	err := s.Pool.QueryRow(ctx, query, base, target).Scan(&rate, &fetchedAt)
	if err != nil {
		return 0, false, err
	}

	isStale := time.Since(fetchedAt) > CacheDuration
	return rate, isStale, nil
}

func (s *ExchangeRateService) fetchFromAPI(ctx context.Context, base, target string) (float64, error) {
	url := fmt.Sprintf("https://api.frankfurter.app/latest?from=%s&to=%s", base, target)

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var result struct {
		Rates map[string]float64 `json:"rates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	rate, exists := result.Rates[target]
	if !exists {
		return 0, fmt.Errorf("rate not found for %s", target)
	}
	return rate, nil
}

func (s *ExchangeRateService) updateCache(ctx context.Context, base, target string, rate float64) {
	query := `
	INSERT INTO exchange_rates_cache(
		base_currency, target_currency, rate, fetched_at) 
	VALUES($1, $2, $3, NOW())
	ON CONFLICT(base_currency, target_currency)
	DO UPDATE SET rate = $3, fetched_at = NOW()
	`
	_, err := s.Pool.Exec(ctx, query, base, target, rate)
	if err != nil {
		fmt.Printf("Warning: Failed to cache exchange rate: %v\n", err)
	}
}
