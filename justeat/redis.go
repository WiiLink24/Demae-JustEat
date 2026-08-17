package justeat

import "time"

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
