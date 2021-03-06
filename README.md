# googshorty
Google URL shortener API client

## Requirements
In your environment, please set variable `GOOGLE_URL_SHORTENER_API_KEY`

## Usage
Please see file [example_test.go](./example_test.go) or follow below

* Preamble:
```go
package main 

import (
	"fmt"
	"log"

	"github.com/orijtech/googshorty/v1"
)
```

* Shorten a URL
```go
func main() {
	client, err := googshorty.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	details, err := client.Shorten("https://github.com/orijtech/googshorty")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ShortURL: %q\n", details.ShortURL)
}
```

* Expand a shortened URL
```go
func main() {
	client, err := googshorty.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	details, err := client.Expand("https://goo.gl/XRdHKo")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("LongURL: %q\n", details.LongURL)
}
```

* Lookup analytics for a shortened URL
```go
func main() {
	client, err := googshorty.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	details, err := client.LookupAnalytics("https://goo.gl/XRdHKo")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Analytics: %#v\n", details)
	if analytics := details.Analytics; analytics != nil {
		fmt.Printf("AllTime: %#v\n", analytics.AllTime)
		fmt.Printf("In Last Month: %#v\n", analytics.AllTime)
		fmt.Printf("In Last Week: %#v\n", analytics.WithinLastWeek)
		fmt.Printf("In Last Day: %#v\n", analytics.WithinLastDay)
		fmt.Printf("In Last Two hours: %#v\n", analytics.WithinLast2Hours)
	}
}
```
