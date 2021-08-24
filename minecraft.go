package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/njhanley/mcproto"
)

func handshakePacket(host string, port uint16) (p mcproto.Packet, err error) {
	var (
		n, m int
		data = make([]byte, (5+5)+5+(5+255)+2+5)
	)

	m, err = mcproto.PutVarInt(data[n:], -1)
	if n += m; err != nil {
		return p, err
	}

	m, err = mcproto.PutString(data[n:], host)
	if n += m; err != nil {
		return p, err
	}

	binary.BigEndian.PutUint16(data[n:], port)
	n += 2

	m, err = mcproto.PutVarInt(data[n:], 1)
	if n += m; err != nil {
		return p, err
	}

	return mcproto.Packet{ID: 0x00, Data: data[:n]}, nil
}

type mcStatus struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"sample"`
	} `json:"players"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
	Favicon []byte `json:"favicon"`
}

func queryMinecraft(addr string, timeout time.Duration) (*mcStatus, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var (
		host string
		port uint16
		n    int
		buf  = make([]byte, (5+5)+(5+32767))
	)

	if _host, _port, err := net.SplitHostPort(addr); err != nil {
		return nil, err
	} else {
		p, err := strconv.ParseUint(_port, 10, 16)
		if err != nil {
			return nil, err
		}
		host, port = _host, uint16(p)
	}

	// handshake
	p, err := handshakePacket(host, port)
	if err != nil {
		return nil, err
	}
	n, err = mcproto.PutPacket(buf, p)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(buf[:n])
	if err != nil {
		return nil, err
	}

	// request
	n, err = mcproto.PutPacket(buf, mcproto.Packet{})
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(buf[:n])
	if err != nil {
		return nil, err
	}

	// response
	n, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}
	p, n, err = mcproto.GetPacket(buf[:n])
	if err != nil {
		return nil, err
	}
	if p.ID != 0x00 {
		return nil, fmt.Errorf("unexpected packet id %#x", p.ID)
	}

	// decode status
	jsonResponse, n, err := mcproto.GetString(p.Data)
	if err != nil {
		return nil, err
	}
	status := new(mcStatus)
	err = json.Unmarshal([]byte(jsonResponse), status)
	if err != nil {
		return nil, err
	}

	return status, nil
}
