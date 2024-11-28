package ws

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"sync/atomic"
	"time"

	"github.com/docker/docker/api/types/container"

	"github.com/docker/docker/client"
	"github.com/gofiber/contrib/websocket"
)

type KeepAliveMessage struct {
	Interval int64 `json:"interval"`
}

func LogStreamHandler(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	containerId := c.Params("id")
	log.Printf("Log listener connected for %s\n", containerId)

	// Create Docker client
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error creating Docker client: %v", err)
		return
	}

	hdr := make([]byte, 8)
	i, err := cli.ContainerLogs(ctx, containerId, container.LogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: false,
		Follow:     true,
		Tail:       "40",
	})
	if err != nil {
		log.Printf(err.Error())
		return
	}

	var keepAliveRemaining int64
	atomic.StoreInt64(&keepAliveRemaining, 5)
	c.SetReadDeadline(time.Now().Add(5 * time.Second))

	go func() {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			keepAliveMessage := KeepAliveMessage{}
			err = json.Unmarshal(msg, &keepAliveMessage)
			if err != nil {
				log.Printf("Error unmarshalling keep alive message: %v", err)
				continue
			}
			interval := min(keepAliveMessage.Interval, 5)
			atomic.StoreInt64(&keepAliveRemaining, interval)
			c.SetReadDeadline(
				time.Now().Add(time.Duration(interval) * time.Second),
			)
		}
		log.Printf("Read loop closed\n")
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				atomic.AddInt64(&keepAliveRemaining, -1)
			default:
				if keepAliveRemaining <= 0 {
					cancel()
					return
				}
			}
		}

	}()

	for {
		_, err := i.Read(hdr)
		if errors.Is(err, context.Canceled) {
			break
		} else if err != nil {
			log.Printf("Error in REPL loop")
			log.Printf(err.Error())
			cancel()
			break
		}
		count := binary.BigEndian.Uint32(hdr[4:])
		dat := make([]byte, count)
		_, err = i.Read(dat)
		c.WriteMessage(websocket.TextMessage, dat)
	}
	log.Printf("Log listener disconnected for %s\n", containerId)
}
