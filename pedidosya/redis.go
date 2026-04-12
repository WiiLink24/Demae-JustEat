package justeat

import "github.com/WiiLink24/DemaeJustEat/logger"

// SetKey sets a key with a value in redis if it does not already exist.
func (j *JEClient) SetKey(key string, value string) error {
	err := j.rdb.SetNX(j.Context, key, value, 0).Err()
	if err != nil {
		return err
	}

	logger.Debug("REDIS", "SetKey", key, value)

	return j.rdb.SetNX(j.Context, value, key, 0).Err()
}

func (j *JEClient) GetKey(key string) (string, error) {
	return j.rdb.Get(j.Context, key).Result()
}

func (j *JEClient) KeyExists(key string) bool {
	return j.rdb.Exists(j.Context, key).Val() != 0
}
