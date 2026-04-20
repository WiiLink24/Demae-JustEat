package justeat

// SetKey sets a key with a value in redis if it does not already exist.
func (j *JEClient) SetKey(key string, value string) error {
	return j.rdb.MSetNX(j.Context, key, value, value, key).Err()
}

func (j *JEClient) GetKey(key string) (string, error) {
	return j.rdb.Get(j.Context, key).Result()
}

func (j *JEClient) KeyExists(key string) bool {
	return j.rdb.Exists(j.Context, key).Val() != 0
}
