# Gopify


[![Go Reference](https://pkg.go.dev/badge/github.com/oussama4/gopify.svg)](https://pkg.go.dev/github.com/oussama4/gopify)

**Gopify** is a simple package for developing Shopify applications in Go.


## Table of Contents

 - [Usage](#usage)
   - [Oauth](#oauth)
	 - [Start oauth process](#start-oauth-process)
	 - [Oauth callback](#oauth-callback)
   - [API calls](#api-calls)
	 - [REST](#rest)
	 - [Graphql](#graphql)
	 - [Rate limiting](#rate-limiting)
   - [Session tokens](#session-tokens)
   - [Verify a Shopify request](#verify-a-shopify-request)
   - [Verify a webhook](#verify-a-webhook)


## Usage
### Oauth
When developing a public or custom Shopify application you need to get an access token using oauth to use Shopify APIs.

#### Start oauth process
The first thing you need to do to use this package is to create a Gopify instance like the following:
```go
app := &gopify.Gopify{
		ApiKey:      "key",
		ApiSecret:   "secret",
		RedirectUrl: "https://example.com/auth/callback",
		Scopes:      []string{"read_products","read_orders"},
	}
```

Add a http handler to trigger the oauth process.

```go
func startOauth(w http.ResponseWriter, r *http.Request) {
    shopName := r.URL.Query().Get("shop")
    authUrl := app.AuthorizationUrl(shopName, "unique token")
    http.Redirect(w, r, authUrl, http.StatusFound)
}
```

#### Oauth callback
After Shopify authenticates your app, it will send a request to the redirect url that you provided to `gopify.Gopify{}` above. Now you can obtain an access token using `AccessToken` method.

```go
func oauthCallback(w http.ResponseWriter, r *http.Request) {
	shopName := r.URL.Query().Get("shop")
	code := r.URL.Query().Get("code")
	token, err := app.AccessToken(shopName, code)

	// Do something with the token, like querying shopify API.
	...

	// redirect to your application home page
	http.Redirect(w, r, "app url", http.StatusFound)
}
```


### API calls
We can make calls to both Shopify APIs, REST and Graphql using the `Client` object provided by this package.

```go
client := gopify.NewClient("example.myshopify.com", "access token")
```

We can can also pass other options to NewClient like the API version, http timeout, ...

```go
// We can use WithVersion to specify which API version
client := gopify.NewClient("example.myshopify.com", "access token", gopify.WithVersion("2021-10"))

// Use WithTimeout to set a custom http timeout instead 10 seconds
client := gopify.NewClient("example.myshopify.com", "access token", WithTimeout(20))
```

#### REST
```go
// Perform a Get request
products, err := client.Get("products.json")

// Perform a Post request
_, err := client.Post("products.json", gopify.Body{})
```

#### Graphql
To send a Graphql query, we use the `Graphql` method defined in the api `Client` type.

```go
query := `
	{
      products (first: 10) {
        edges {
          node {
            id
            title
          }
        }
      }
    }
`
// the second parameter is for query variables, here we pass nil because we don't have any variables
products, nil := client.Graphql(query, nil)
```

#### Rate limiting
Shopify APIs are rate limited, so if that happens you can use the `WithRetry` option to specify how many times to retry a request.

```go
// retry the request 10 times when hit the rate limit
client := gopify.NewClient("example.myshopify.com", "access token", WithRetry(10))
```

### Session tokens
If you are building an [embedded Shopify app](https://shopify.dev/apps/getting-started/app-types#embedded-apps) then you need to authenticate your app with [session tokens](https://shopify.dev/apps/auth/session-tokens).

This package provides you with facilities for decoding a session token and extracting its payload, and also a way to verify the authenticity of the token.

```go
// decode the token 
payload, err := app.DecodeSessionToken("token")

// verify the signature of the token
err := app.VerifyTokenSignature("token")
```

There is also a higher level way to verify the authenticity of token using the [VerifyToken](https://pkg.go.dev/github.com/oussama4/gopify#Gopify.VerifyToken) http middleware.

### Verify a Shopify request
To verify the authenticity of the request from Shopify we can verify the signature of a hmac parameter included in every request from shopify using [VerifyRequest](https://pkg.go.dev/github.com/oussama4/gopify#Gopify.VerifyRequest) http middleware.

### Verify a webhook
To verify that a webhook request is from Shopify we can use [VerifyWebhook](https://pkg.go.dev/github.com/oussama4/gopify#Gopify.VerifyWebhook) function.