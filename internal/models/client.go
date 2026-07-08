package models

import (
	"net"
	"time"
)

type Client struct {
	ID       string
	Conn     net.Conn
	JoinTime time.Time
}
