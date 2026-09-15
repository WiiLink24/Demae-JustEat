package justeat

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/redis/go-redis/v9"
)

var ErrBasketItemIndexNotFound = errors.New("basket item index not found")

const basketItemIndexTTL = 3 * time.Hour

// basketItemRef is what basketDelete needs to remove a line: its type and Just Eat IDs
type BasketItemRef struct {
	IsDeal           bool     `json:"isDeal"`
	BasketProductIds []string `json:"basketProductIds"`
}

// SetKey sets a key with a value in redis if it does not already exist.
func (j *JEClient) SetKeyReversible(key string, value string) error {
	err := j.rdb.Set(j.Context, key, value, 3*time.Hour).Err()
	if err != nil {
		return err
	}
	return j.rdb.Set(j.Context, value, key, 3*time.Hour).Err()
}

func (j *JEClient) SetKey(key string, value string) error {
	return j.rdb.Set(j.Context, key, value, 3*time.Hour).Err()
}

func (j *JEClient) GetKey(key string) (string, error) {
	return j.rdb.GetEx(j.Context, key, 3*time.Hour).Result()
}

func (j *JEClient) KeyExists(key string) bool {
	return j.rdb.Exists(j.Context, key).Val() != 0
}

// ShortenID handles IDs demae.CompressUUID can't and stays stable across repeat calls for the same id
func (j *JEClient) ShortenID(id string) (string, error) {
	if j.KeyExists(id) {
		return j.GetKey(id)
	}

	short := demae.CompressUUID(demae.UUID())
	_, err := j.rdb.SetArgs(j.Context, id, short, redis.SetArgs{Mode: "NX", TTL: 3 * time.Hour}).Result()
	if errors.Is(err, redis.Nil) {
		return j.GetKey(id)
	} else if err != nil {
		return "", err
	}

	if err := j.rdb.Set(j.Context, short, id, 3*time.Hour).Err(); err != nil {
		return "", err
	}

	return short, nil
}

func (j *JEClient) SetBasketItems(basketId string, refs []BasketItemRef) error {
	value, err := json.Marshal(refs)
	if err != nil {
		return err
	}

	return j.rdb.Set(j.Context, "items:"+basketId, value, basketItemIndexTTL).Err()
}

func (j *JEClient) GetBasketItemIndex(basketId string, index string) (BasketItemRef, error) {
	i, err := strconv.Atoi(index)
	if err != nil {
		return BasketItemRef{}, ErrBasketItemIndexNotFound
	}

	raw, err := j.rdb.Get(j.Context, "items:"+basketId).Result()
	if errors.Is(err, redis.Nil) {
		return BasketItemRef{}, ErrBasketItemIndexNotFound
	} else if err != nil {
		return BasketItemRef{}, err
	}

	var refs []BasketItemRef
	if err := json.Unmarshal([]byte(raw), &refs); err != nil {
		return BasketItemRef{}, err
	}

	if i < 0 || i >= len(refs) {
		return BasketItemRef{}, ErrBasketItemIndexNotFound
	}

	return refs[i], nil
}
