package config

func NewTestConfigWithYaml(data []byte) (*Config, error) {
	c := &Config{}
	if err := c.parse(data); err != nil {
		return nil, err
	}

	return c, nil
}
