package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (c *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *RedisCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.SetNX(ctx, key, data, ttl).Result()
}

// Cache key generators
const (
	KeyPrefixArticle        = "article:"
	KeyPrefixArticleSlug    = "article:slug:"
	KeyPrefixArticleList    = "articles:list:"
	KeyPrefixTrending       = "articles:trending"
	KeyPrefixCategory       = "category:"
	KeyPrefixCategories     = "categories:all"
	KeyPrefixPolitician     = "politician:"
	KeyPrefixPoliticianSlug = "politician:slug:"
	KeyPrefixPoliticians    = "politicians:all"
	KeyPrefixPoliticianList = "politicians:list:"
	KeyPrefixRateLimit      = "ratelimit:"

	// Location cache keys
	KeyPrefixRegion            = "region:"
	KeyPrefixRegionSlug        = "region:slug:"
	KeyPrefixRegions           = "regions:all"
	KeyPrefixProvince          = "province:"
	KeyPrefixProvinceSlug      = "province:slug:"
	KeyPrefixProvinces         = "provinces:"
	KeyPrefixAllProvinces      = "provinces:all"
	KeyPrefixCity              = "city:"
	KeyPrefixCitySlug          = "city:slug:"
	KeyPrefixCities            = "cities:"
	KeyPrefixBarangay          = "barangay:"
	KeyPrefixBarangaySlug      = "barangay:slug:"
	KeyPrefixBarangays         = "barangays:"
	KeyPrefixDistrict          = "district:"
	KeyPrefixLocationHierarchy = "location:hierarchy:"
)

func ArticleKey(id string) string {
	return KeyPrefixArticle + id
}

func ArticleSlugKey(slug string) string {
	return KeyPrefixArticleSlug + slug
}

func ArticleListKey(page, perPage int, filter string) string {
	return fmt.Sprintf("%s%d:%d:%s", KeyPrefixArticleList, page, perPage, filter)
}

func TrendingKey() string {
	return KeyPrefixTrending
}

func CategoryKey(id string) string {
	return KeyPrefixCategory + id
}

func CategoriesKey() string {
	return KeyPrefixCategories
}

func RateLimitKey(ip string) string {
	return KeyPrefixRateLimit + ip
}

func PoliticianKey(id string) string {
	return KeyPrefixPolitician + id
}

func PoliticianSlugKey(slug string) string {
	return KeyPrefixPoliticianSlug + slug
}

func PoliticiansKey() string {
	return KeyPrefixPoliticians
}

func PoliticianListKey(page, perPage int, filter string) string {
	return fmt.Sprintf("%s%d:%d:%s", KeyPrefixPoliticianList, page, perPage, filter)
}

// Location cache key functions
func RegionKey(id string) string {
	return KeyPrefixRegion + id
}

func RegionSlugKey(slug string) string {
	return KeyPrefixRegionSlug + slug
}

func RegionsKey() string {
	return KeyPrefixRegions
}

func ProvinceKey(id string) string {
	return KeyPrefixProvince + id
}

func ProvinceSlugKey(slug string) string {
	return KeyPrefixProvinceSlug + slug
}

func ProvincesKey(regionID string) string {
	return KeyPrefixProvinces + regionID
}

func AllProvincesKey() string {
	return KeyPrefixAllProvinces
}

func CityKey(id string) string {
	return KeyPrefixCity + id
}

func CitySlugKey(slug string) string {
	return KeyPrefixCitySlug + slug
}

func CitiesKey(provinceID string) string {
	return KeyPrefixCities + provinceID
}

func BarangayKey(id string) string {
	return KeyPrefixBarangay + id
}

func BarangaySlugKey(slug string) string {
	return KeyPrefixBarangaySlug + slug
}

func BarangaysKey(cityID string) string {
	return KeyPrefixBarangays + cityID
}

func DistrictKey(id string) string {
	return KeyPrefixDistrict + id
}

func LocationHierarchyKey(barangayID string) string {
	return KeyPrefixLocationHierarchy + barangayID
}
