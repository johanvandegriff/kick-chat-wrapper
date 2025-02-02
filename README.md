# Introduction

`kickchatwrapper` is a Go package that provides a wrapper for interacting with the Kick websocket. It allows you to subscribe to chat channels and receive chat messages.

# Installation

```
go get github.com/johanvandegriff/kick-chat-wrapper
```

# Usage

```golang
package main

import (
    "fmt"
    "time"

    kickchatwrapper "github.com/johanvandegriff/kick-chat-wrapper"
)

func main() {
	//if this library ever stops working, it's probably because kick changed the websocket URL
	//you can find the new URL by going to the network tab on your kick page, refresh, and type "pusher" in the filter box
	//then you can manually set the URL with the following:
	kickchatwrapper.APIURL = "wss://ws-us2.pusher.com/app/32cbd69e4b950bf97679?protocol=7&client=js&version=8.4.0-rc2&flash=false"

    client, err := kickchatwrapper.NewClient()
    if err != nil {
        // handle error
    }

    client.JoinChannelByID(15034797) //ID for https://kick.com/JJVanVan
    //find your ID here: https://kick.com/api/v2/channels/JJVanVan
    //search in the page for "chatroom":{"id":

    messageChan := client.ListenForMessages()

    go func() {
    for message := range messageChan {
        fmt.Printf("Received chat message: %+v\n", message)
    }
    }()

    defer client.Close()

    for {
        time.Sleep(1 * time.Second)
    }
}
```

# Notes

Right now it is only possible to join chat room using user ID because to be able to join it by username we would need to call their API but its protected by CloudFlare.
