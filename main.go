package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"layeh.com/gumble/gumble"
	"net"
	"net/http"
	"time"
)

type BotConfig struct {
	addr      string
	channel   string
	webhook   string
	threshold int
	interval  int
}

func main() {
	BFConfig := new(BotConfig)
	flag.StringVar(&BFConfig.addr, "addr", "", "Mumble server address")
	flag.StringVar(&BFConfig.channel, "channel", "", "Channel to watch")
	flag.StringVar(&BFConfig.webhook, "webhook", "", "Webhook to POST to")
	flag.IntVar(&BFConfig.threshold, "threshold", 4, "Minimum required players")
	flag.IntVar(&BFConfig.interval, "interval", 90, "Seconds to wait")
	flag.Parse()

	// Create client
	GumbleConfig := gumble.NewConfig()
	GumbleConfig.Username = "BOT"
	TLSConfig := new(tls.Config)
	TLSConfig.InsecureSkipVerify = true

	var err error
	conn, err := gumble.DialWithDialer(new(net.Dialer), BFConfig.addr, GumbleConfig, TLSConfig)
	if err != nil {
		panic(err)
	}
	conn.Self.SetSelfMuted(true)
	conn.Self.SetSelfDeafened(true)
	target := conn.Channels.Find(BFConfig.channel)

	// Start watching
	active := false
	for {
		total := len(target.Users)

		if !active && total >= BFConfig.threshold {
			active = true
			fmt.Printf("%d: %d people are playing, notifying\n", time.Now().Unix(), total)
			message := map[string]string{"content": fmt.Sprintf("There's %d playing Battlefield right now. Get on mumble if you want in.", total)}
			jb, _ := json.Marshal(message)
			_, _ = http.Post(BFConfig.webhook, "application/json", bytes.NewBuffer(jb))
		} else if active && total >= BFConfig.threshold {
			fmt.Printf("%d: Active session, %d playing\n", time.Now().Unix(), total)
		} else {
			fmt.Printf("%d: %d playing, need %d more\n", time.Now().Unix(), total, BFConfig.threshold-total)
			active = false
		}

		fmt.Printf("%d: Waiting %d seconds\n", time.Now().Unix(), BFConfig.interval)
		time.Sleep(time.Duration(BFConfig.interval) * time.Second)
	}
}
