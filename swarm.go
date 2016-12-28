package swarm

import (
	"crypto/sha1"
	"encoding/json"
	"time"
)

var InBox map[string]Message
var Received [][]byte

type Cont []byte

type Message struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Content []byte `json:"content"`
	Id      []byte `json:"id"`
	Sent    int64  `json:"sent"`
}

func (n *Node) Write(message []byte, title ...string) bool {
	var err error

	m := new(Message)
	m.Content = message
	if len(title) > 0 {
		m.Title = title[0]
	}
	m.Type = "regular"
	sha := sha1.New()
	sha.Write([]byte(time.Now().String()))
	m.Id = sha.Sum(nil)
	m.Sent = time.Now().Unix()

	js, err := json.Marshal(m)

	if err != nil {
		print(err.Error())
		return false
	}

	for key, value := range LocalNode.Children {
		_, err = value.conn.Write(js)
		if err != nil {
			println(err.Error())
			LocalNode.Children = append(LocalNode.Children[:key], LocalNode.Children[key+1:]...)
		}
	}
	if RemoteNode != nil {
		_, err = RemoteNode.conn.Write(js)
	}

	LocalNode.HandleMessage(js, LocalNode)

	if err != nil {
		println(err.Error())
		return false
	}

	return true
}
