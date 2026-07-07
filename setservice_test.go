package Apple

import (
	"sync"
	"testing"
)

// TestSetService_ConcurrentIsRaceFree guards the data race in SetService:
// concurrently pointing one Client at different services must not corrupt
// shared state. Run under `go test -race` to detect regressions.
func TestSetService_ConcurrentIsRaceFree(t *testing.T) {
	c := NewClient(false, "KID", "ISS", "BID", testPrivateKeyPEM(t))
	services := []AppleClient{
		AppStoreConnectClient,
		AppStoreServerClient,
		AppStoreServerNotificationsClient,
	}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.SetService(services[i%len(services)])
		}(i)
	}
	wg.Wait()
}
