// main.go  –  orderbook-snapshot-diff
// go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type snapshot struct {
	Bids map[float64]float64 `json:"bids"`
	Asks map[float64]float64 `json:"asks"`
}

type l2update struct {
	Changes [][]string `json:"changes"`
}

type msg struct {
	Type      string    `json:"type"`
	ProductID string    `json:"product_id"`
	Snapshot  *snapshot `json:"snapshot,omitempty"`
	Updates   *l2update `json:"changes,omitempty"`
	Time      string    `json:"time"`
}

// lock-free hot-path: just maps
var (
	book = struct {
		sync.RWMutex
		bids map[float64]float64
		asks map[float64]float64
	}{bids: make(map[float64]float64), asks: make(map[float64]float64)}
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// 1) subscribe to full L2
	conn, _, err := websocket.DefaultDialer.DialContext(ctx,
		"wss://ws-feed.exchange.coinbase.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	sub := map[string]any{
		"type":        "subscribe",
		"product_ids": []string{"BTC-USD"},
		"channels":    []string{"level2"},
	}
	if err := conn.WriteJSON(sub); err != nil {
		log.Fatal(err)
	}

	// 2) stream & update
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		var m msg
		if err := conn.ReadJSON(&m); err != nil {
			log.Printf("read: %v", err)
			continue
		}
		switch m.Type {
		case "snapshot":
			loadSnap(m.Snapshot)
		case "l2update":
			apply(m.Updates, m.Time)
		}
	}
}

func loadSnap(s *snapshot) {
	book.Lock()
	defer book.Unlock()
	for _, b := range s.Bids {
		book.bids[b] = b
	}
	for _, a := range s.Asks {
		book.asks[a] = a
	}
}

func apply(u *l2update, ts string) {
	book.Lock()
	defer book.Unlock()
	for _, ch := range u.Changes {
		side, pxStr, szStr := ch[0], ch[1], ch[2]
		px := strToF(pxStr)
		sz := strToF(szStr)
		if side == "buy" {
			if sz == 0 {
				delete(book.bids, px)
			} else {
				book.bids[px] = sz
			}
		} else {
			if sz == 0 {
				delete(book.asks, px)
			} else {
				book.asks[px] = sz
			}
		}
	}
	// cross detector
	bestBid := maxKey(book.bids)
	bestAsk := minKey(book.asks)
	if bestBid >= bestAsk && bestBid > 0 && bestAsk > 0 {
		t, _ := time.Parse(time.RFC3339, ts)
		fmt.Printf("cross %d %f %f\n", t.UnixNano(), bestBid, bestAsk)
	}
}

// helpers
func strToF(s string) float64 {
	var f float64
	fmt.Sscan(s, &f)
	return f
}
func maxKey(m map[float64]float64) float64 {
	max := 0.0
	for k := range m {
		if k > max {
			max = k
		}
	}
	return max
}
func minKey(m map[float64]float64) float64 {
	min := 1e12
	for k := range m {
		if k < min {
			min = k
		}
	}
	return min
}
