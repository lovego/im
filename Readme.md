# im
Package im implement instant messages by long pull.
Multiple node is supported by redis publish/subscribe mechanism.

[![GoDoc](https://godoc.org/github.com/lovego/im?status.svg)](https://godoc.org/github.com/lovego/im)

```go
var im = New("redis://@localhost/0", "im", nil, logger.New(os.Stderr))

// Push a new version of business "notify" for user "bob" in "demo" system.
if err := im.Push("demo", []string{"bob"}, "notify"); err != nil {
    log.Panic(err)
}

// Pull new versions of businesses "notify" and "chats" for user "bob" in "demo" system.
// It blocked until get a new version of the businesses or reach the one second timeout.
// Only businesses that have a new version is returned in the map,
// an empty map is returned if no businesses has a new version.
if versions, err := im.Pull("demo", "bob", map[string]string{
    "notify": "1", "chats": ""
}, time.Second); err != nil {
    log.Panic(err)
} else {
    // prints something like: map[notify:2 chats:1]
    fmt.Println(versions)
}
```

