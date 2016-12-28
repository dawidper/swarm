package swarm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	LocalNode  Node
	RemoteNode *Node
)

func (remote Node) HandleFail() {
	LocalNode.Connect(remote.Parent[0])
}

func (local *Node) Connect(remote Node) {
	dial, err := net.Dial("tcp", remote.Ip+":"+strconv.Itoa(remote.Port))
	println("Connecting to " + remote.Ip + ":" + strconv.Itoa(remote.Port))
	if err != nil {
		println(err.Error())
		return
	}

	dial.Write([]byte(remote.Name + "\n"))
	RemoteNode = &remote
	RemoteNode.conn = dial

	LocalNode.Parent = append(LocalNode.Parent, *RemoteNode)
	println("Connected to " + RemoteNode.Name)

	var children Nodes

	for _, v := range local.Children {
		if v.Name != remote.Name {
			children = append(children, v)
		}
	}
	LocalNode.Children = children
}

func (local *Node) AddChild(remote Node) {
	remote.conn.Write([]byte(remote.Ip))
	for _, value := range local.Children {
		enc, _ := json.Marshal(value)
		remote.conn.Write(enc)
	}
	LocalNode.Children = append(LocalNode.Children, remote)
	go remote.WaitFormessage()
}

func (node *Node) WaitFormessage() {
	println("Waiting for message")
	for {
		message, err := bufio.NewReader(node.conn).ReadBytes('\n')
		if err != nil {
			if node.Name == string(message) {
				RemoteNode.HandleFail()
			}
			go LocalNode.ResendMessage(message, *node)
			go LocalNode.HandleMessage(message, *node)
		}

	}
}

func (node *Node) ResendMessage(m []byte, sender Node) {
	for _, v := range LocalNode.Children {
		if v.Name != sender.Name {
			v.conn.Write(m)
		}
	}

	if RemoteNode != nil {
		if RemoteNode.Name != sender.Name {
			RemoteNode.conn.Write(m)
		}
	}
}

func (node *Node) HandleMessage(text []byte, sender Node) {
	var message Message
	err := json.Unmarshal(text, message)
	if err != nil {
		Received = append(Received, text)
		return
	}
	fmt.Println("Message from", sender.Name)
	switch message.Type {
	case "system":
		{
			LocalNode.HandleSystemMessage(message)
		}
	case "regular":
		{
			key := string(message.Id)
			InBox[key] = message
		}
	}

}

func (node *Node) HandleSystemMessage(m Message) {
	switch m.Title {
	case "UpdateChildren":
		{
			var children Nodes
			err := json.Unmarshal(m.Content, &children)
			if err != nil {
				return
			}
			node.Children = children
		}
	}
}

func (local *Node) HandleNewConnection(conn net.Conn) {
	fmt.Println(conn.RemoteAddr().String())

	name, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	name = strings.Trim(name, "\r\n")

	conn.Write([]byte("Hello my minion " + name))
	remote := new(Node)
	s := strings.Split(conn.RemoteAddr().String(), ":")
	remote.Ip = s[0]
	remote.Port, err = strconv.Atoi(s[1])
	if err != nil {
		return
	}
	remote.Name = name
	remote.conn = conn
	local.AddChild(*remote)
}

func (local *Node) Start(server_name, server_ip string, server_port int) {
	LocalNode = *local
	l, err := net.Listen("tcp", getListen())
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	println("Listening on " + getListen() + " as " + getNodeName())
	if os.Getenv("server_ip") != "" {

		remote := NewNode(server_name, server_ip, server_port)
		LocalNode.Connect(remote)
	}
	for {
		println("STARTING LISTENER")
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go local.HandleNewConnection(conn)
	}
}

func getListen() string {
	return os.Getenv("local_ip") + ":" + strconv.Itoa(LocalNode.Port)
}

func getListenPort() int {
	port, _ := strconv.Atoi(os.Getenv("local_port"))
	return port
}

func getNodeName() string {
	return os.Getenv("local_name")
}
