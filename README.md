# go-typed-singleflight

Generic-supporting [golang.org/x/sync/singleflight](https://golang.org/x/sync/singleflight).

Example usage ([Go Playground](https://go.dev/play/p/KM383MCGJGh)):

```go
package main

import (
	"fmt"
	"sync"
	"time"

	typedsf "github.com/t2bot/go-typed-singleflight"
)

type MyValue string

func main() {
	g := new(typedsf.Group[MyValue])

	workFn := func() (MyValue, error) {
		// for example purposes only, sleep for a moment
		time.Sleep(1 * time.Second)
		return MyValue("this is a value"), nil
	}

	key := "my_resource" // deduplication key

	// This loop simulates multiple requests
	wg := new(sync.WaitGroup)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err, shared := g.Do(key, workFn)
			if err != nil {
				panic(err)
			}

			if shared {
				fmt.Println("Response was shared!")
				// When true, the workFn was only called once and its output used
				// multiple times.
			} else {
				// This shouldn't happen in this example
				fmt.Println("WARN: Response was not shared!")
			}

			fmt.Println("Got val: ", val)
		}
	}

	fmt.Println("Waiting for all requests to finish")
	wg.Wait()
	fmt.Println("Done!")
}
```
