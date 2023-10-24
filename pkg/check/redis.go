package check

type RedisCheck struct {
}

func (r RedisCheck) SetUp() error {
	return nil
}

func (r RedisCheck) Check() (int, error) {
	return 0, nil
}

func (r RedisCheck) Validate() error {
	return nil
}