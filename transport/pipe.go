package transport

import (
	"io"
	"net"
	"sync"
)

// pipe copies data bidirectionally between two connections.
// Returns when either connection is closed or an error occurs.
func pipe(a, b net.Conn) error {
	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 2)

	// Copy from a to b
	go func() {
		defer wg.Done()
		if _, err := io.Copy(b, a); err != nil {
			errChan <- err
		}
	}()

	// Copy from b to a
	go func() {
		defer wg.Done()
		if _, err := io.Copy(a, b); err != nil {
			errChan <- err
		}
	}()

	// Wait for both copies to complete
	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
