# Photon-controller-go-SDK

# WARNING: Photon Controller GO SDK is no longer actively maintained by VMware.

VMware has made the difficult decision to stop driving this project and therefore we will no longer actively respond 
to issues or pull requests. If you would like to take over maintaining this project independently from VMware, please 
let us know so we can add a link to your forked project here.

Thank You.

# Getting Started with Photon Controller SDK

1. If you haven't already, set up a Go workspace according to the
   [Go docs](http://golang.org/doc).
2. Install the Go SDK. Normally this is done with "go get".
3. Setup GOPATH environment variable

	Then:
	```
	mkdir -p $GOPATH/src/github.com/vmware
	cd $GOPATH/src/github.com/vmware
	git clone (github.com/vmware or gerrit)/photon-controller-go-sdk
	```

# Testing for developers

For our Go SDK, we use Ginkgo and Gomega testing framework.
Install Gingko and Gomega:
```
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega
```


To run the tests, use the command:
```
go test ./...
```

You can add flags from both go and ginkgo for more detailed debugging.
Some helpful flags:
```
-v                                   Verbose output
--ginkgo.v                           Verbose Ginkgo output (Prints all specs as they begin)
--ginkgo.noColor                     No color output
--ginkgo.slowSpecThreshold=<integer> Ignore Ginkgo warnings of slow tests under threshold seconds
--ginkgo.focus <regex>               Run specs that match the regex
--ginkgo.skip <regex>                Skip specs that match the regex
```

To run the SDK tests against a real deployment, you can set the following environment variables:
```
TEST_ENDPOINT Photon Controller Endpoint
REAL_AGENT    (optional)

Photon Controller credentials in the form of either:
API_ACCESS_TOKEN
or
USERNAME
PASSWORD
```

Note: Some tests are skipped when run against a real deployment.


**WARNING: Currently some tests are known to fail when run against a real deployment.**

To run the Lightwave portion of SDK tests against a real Lightwave set the following environment variables:
```
LIGHTWAVE_ENDPOINT
LIGHTWAVE_USERNAME
LIGHTWAVE_PASSWORD
```

## Sample App

Here's a quick sample app that will retrieve Photon Controller status from a
[local devbox].
In this example, it's under $GOPATH/src/sdkexample/main.go:

```golang
package main

import (
	"fmt"
	"github.com/vmware/photon-controller-go-sdk/photon"
	"log"
)

func main() {
	clientOptions := photon.ClientOptions{IgnoreCertificate: true}
	client := photon.NewClient("https://localhost:9000", &clientOptions, nil)
	tokenOptions, err := client.Auth.GetTokensByPassword("username", "password")
	if err != nil {
		log.Fatal(err)
	}
	clientOptions = photon.ClientOptions{IgnoreCertificate: true, TokenOptions: tokenOptions}
	client = photon.NewClient("https://localhost:9000", &clientOptions, nil)
	status, err := client.System.GetSystemStatus()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(status)
}
```

Then build it and run it:

```
cd $GOPATH/src/sdkexample
go build ./...
./sdkexample
```

And the output should look something like this:
`&{READY [{PHOTON_CONTROLLER  READY}]}`

## Using APIs that return tasks

Most Photon Controller APIs use a task model. The API will return a task object,
which will indicate the state of the task (such as queued, completed, error, etc).
These tasks return immediately and the caller must poll to find out when the task
has been completed.The Go SDK provides a tasks API to do this for you,
with built-in retry and error handling.

Let's expand the sample app to create a new tenant:

```
package main

import (
	"fmt"
	"github.com/vmware/photon-controller-go-sdk/photon"
	"log"
)

func main() {
	clientOptions := photon.ClientOptions{IgnoreCertificate: true}
	client := photon.NewClient("https://localhost:9000", &clientOptions, nil)
	tokenOptions, err := client.Auth.GetTokensByPassword("username", "password")
	if err != nil {
		log.Fatal(err)
	}
	clientOptions = photon.ClientOptions{IgnoreCertificate: true, TokenOptions: tokenOptions}
	client = photon.NewClient("https://localhost:9000", &clientOptions, nil)
	status, err := client.System.GetSystemStatus()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(status)

	// Let's create a new tenant
	tenantSpec := &photon.TenantCreateSpec{Name: "new-tenant"}

	task, err := client.Tenants.Create(tenantSpec)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for task completion
	task, err = client.Tasks.Wait(task.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID of new tenant is: %s\n", task.Entity.ID)
}

```

It should now output this:

```
&{READY [{PHOTON_CONTROLLER  READY}]}
ID of new tenant is: c8989a40-0fa4-4d9a-8e73-2fe4d28d0065
```
