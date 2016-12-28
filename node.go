package swarm

import "net"

type Node struct {
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	Children Nodes  `json:"children"`
	Parent   Nodes  `json:"parent"`
	conn     net.Conn
}

type Nodes []Node

func NewNode(name, ip string, port int) Node {
	node := new(Node)
	node.Ip = ip
	node.Name = name
	node.Port = port
	return *node
}
