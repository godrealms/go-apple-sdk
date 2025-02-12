package Apple

import "time"

type Config struct {
	BaseUrl          string        // Request address
	Timeout          time.Duration // Request timeout
	RetryCount       int           // Number of retry times
	RetryWaitTime    time.Duration // Retry waiting time
	RetryMaxWaitTime time.Duration // Retry maximum waiting time
	Kid              string        // Your private key ID from App Store Connect (Ex: 2X9R4HXF34)
	Iss              string        // Your issuer ID from the Keys page in App Store Connect (Ex: “57246542-96fe-1a63-e053-0824d011072a")
	Bid              string        // Your app’s bundle ID (Ex: “com.example.testbundleid”)
}

func NewConfig(baseUrl, kid, iss, bid string) *Config {
	return &Config{
		BaseUrl:          baseUrl,
		Timeout:          10 * time.Second,
		RetryCount:       3,
		RetryWaitTime:    3 * time.Second,
		RetryMaxWaitTime: 10 * time.Second,
		Kid:              kid,
		Iss:              iss,
		Bid:              bid,
	}
}

func (c *Config) SetWithTimeout(timeout time.Duration) {
	c.Timeout = timeout
}
func (c *Config) SetWithRetryCount(count int) {
	c.RetryCount = count
}
func (c *Config) SetWithRetryWaitTime(waitTime time.Duration) {
	c.RetryWaitTime = waitTime
}
func (c *Config) SetWithRetryMaxWaitTime(waitTime time.Duration) {
	c.RetryMaxWaitTime = waitTime
}
