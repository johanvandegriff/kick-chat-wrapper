package kickchatwrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var (
	APIURL = "wss://ws-us2.pusher.com/app/32cbd69e4b950bf97679?protocol=7&client=js&version=8.4.0-rc2&flash=false"
)

type Client struct {
	ws             *websocket.Conn
	joinedChannels map[int]bool
	quit           chan bool
	debug          bool
}

type PusherSubscribe struct {
	Event string `json:"event"`
	Data  struct {
		Channel string `json:"channel"`
		Auth    string `json:"auth"`
	} `json:"data"`
}

type ChatMessageEvent struct {
	Event   string `json:"event"`
	Data    string `json:"data"`
	Channel string `json:"channel"`
}

type ChatMessage struct {
	ID         string    `json:"id"`
	ChatroomID int       `json:"chatroom_id"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
	Sender     Sender    `json:"sender"`
}

type Sender struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Slug     string   `json:"slug"`
	Identity Identity `json:"identity"`
}

type Identity struct {
	Color  string  `json:"color"`
	Badges []Badge `json:"badges"`
}

type Badge struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Count int    `json:"count"`
}

func (client *Client) printLog(msg string) {
	if !client.debug {
		return
	}
	fmt.Println(msg)
}

func NewClient() (*Client, error) {
	ws, _, err := websocket.DefaultDialer.Dial(APIURL, nil)
	if err != nil {
		return &Client{}, err
	}

	client := &Client{
		ws:             ws,
		joinedChannels: make(map[int]bool),
		quit:           make(chan bool),
		debug:          false,
	}
	return client, err
}

func (client *Client) reconnect() error {
	client.printLog("Reconnecting...")
	client.ws.Close()

	ws, _, dialErr := websocket.DefaultDialer.Dial(APIURL, nil)
	if dialErr != nil {
		return dialErr
	}

	previousChannels := client.joinedChannels
	client.ws = ws
	client.joinedChannels = make(map[int]bool)

	for id := range previousChannels {
		joinErr := client.JoinChannelByID(id)
		if joinErr != nil {
			return joinErr
		}
	}

	return nil
}

func (client *Client) ListenForMessages() <-chan ChatMessage {
	ch := make(chan ChatMessage)
	go func() {
		for {
			select {
			case <-client.quit:
				return
			default:
				_, msg, err := client.ws.ReadMessage()
				if err != nil {
					client.printLog("Error reading message: " + err.Error())
					reconnectErr := client.reconnect()
					if reconnectErr != nil {
						client.printLog("Error reconnecting: " + reconnectErr.Error())
						time.Sleep(time.Second * 5)
					} else {
						client.printLog("Reconnected.")
					}
					continue
				}

				var chatMessageEvent ChatMessageEvent
				errMarshalEvent := json.Unmarshal([]byte(msg), &chatMessageEvent)
				if errMarshalEvent != nil {
					continue
				}

				var chatMessage ChatMessage
				errMarshalMessage := json.Unmarshal([]byte(chatMessageEvent.Data), &chatMessage)
				if errMarshalMessage != nil {
					continue
				}

				ch <- chatMessage
			}
		}
	}()
	return ch
}

func (client *Client) JoinChannelByID(id int) error {
	client.printLog("Joining channel: " + strconv.Itoa(id))
	if _, ok := client.joinedChannels[id]; ok {
		return nil
	}

	pusherSubscribe := PusherSubscribe{
		Event: "pusher:subscribe",
		Data: struct {
			Channel string `json:"channel"`
			Auth    string `json:"auth"`
		}{
			Channel: "chatrooms." + strconv.Itoa(id) + ".v2",
			Auth:    "",
		},
	}

	msg, marshalErr := json.Marshal(pusherSubscribe)
	if marshalErr != nil {
		return errors.New("marshal error")
	}

	err := client.ws.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		return errors.New("error joining channel")
	}

	client.joinedChannels[id] = true
	return nil
}

func (client *Client) SetDebug(debug bool) {
	client.debug = debug
}

func (client *Client) Close() {
	client.quit <- true
	client.ws.Close()
}
