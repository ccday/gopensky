# Gopensky

Go bindings for the OpenSky REST API.

Example: 

```go
// Get all state vectors
api := gopensky.New(&http.Client{})
states, _ := api.Get(&gopensky.Request{})
```
