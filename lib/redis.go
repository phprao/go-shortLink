package lib

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pilu/go-base62"
	"time"
)

const (
	URLIDKEY = "next.url.id"

	// eid -> url
	ShortLinkToUrl = "shortLink:%s:url"

	// urlHash -> eid
	URLHashToIdKey = "urlHash:%s:url"

	// eid -> detail
	ShortLinkDetail = "shortLink:%s:detail"
)

type RedisCli struct {
	cli *redis.Client
}

type URLDetail struct {
	URL                 string        `json:"url"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
	CreatedAt           string        `json:"created_at"`
}

func NewRedisClient(addr string, auth string, db int) *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: auth,
		DB:       db,
	})

	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}

	return &RedisCli{cli: c}
}

// exp minutes
func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	h := ToSha1(url)
	// 查看是否存在
	d, err := r.cli.Get(fmt.Sprintf(URLHashToIdKey, h)).Result()
	if err == redis.Nil {
		// 不存在
	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {
			// 过期了
		} else {
			// 存在则返回短地址
			return d, nil
		}
	}
	// ID自增
	err1 := r.cli.Incr(URLIDKEY).Err()
	if err1 != nil {
		return "", err1
	}
	// 获取ID
	id, err := r.cli.Get(URLIDKEY).Int()
	if err != nil {
		return "", err
	}
	expiration := time.Duration(exp) * time.Minute
	// ID 做 base62 转义
	eid := base62.Encode(id)
	// eid -> url
	err2 := r.cli.Set(fmt.Sprintf(ShortLinkToUrl, eid), url, expiration).Err()
	if err2 != nil {
		return "", err2
	}

	// hash -> eid
	err3 := r.cli.Set(fmt.Sprintf(URLHashToIdKey, h), eid, expiration).Err()
	if err3 != nil {
		return "", err3
	}

	detail, err := json.Marshal(&URLDetail{
		URL:                 url,
		ExpirationInMinutes: expiration,
		CreatedAt:           time.Now().String(),
	})
	if err != nil {
		return "", err
	}

	// eid -> detail
	err4 := r.cli.Set(fmt.Sprintf(ShortLinkDetail, eid), detail, expiration).Err()
	if err4 != nil {
		return "", err4
	}

	return eid, nil
}

func (r *RedisCli) ShortLinkInfo(eid string) (interface{}, error) {
	detail, err := r.cli.Get(fmt.Sprintf(ShortLinkDetail, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, fmt.Errorf("Unknown short URL")}
	}else if err != nil {
		return "", nil
	}else{
		return detail, nil
	}
}

func (r *RedisCli) UnShorten(eid string) (string, error) {
	url, err := r.cli.Get(fmt.Sprintf(ShortLinkToUrl, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, fmt.Errorf("Unknown short URL")}
	}else if err != nil {
		return "", nil
	}else{
		return url, nil
	}
}
