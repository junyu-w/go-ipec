package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/DrakeW/go-ipec/ipec"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	crypto "github.com/libp2p/go-libp2p-crypto"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

func setupLogger() {
	customFormatter := new(log.TextFormatter)
	customFormatter.FullTimestamp = true
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
}

func main() {
	setupLogger()

	peers := flag.String("peers", "", "Peers to connect to")
	port := flag.Int("port", 0, "port to listen to")
	funcFile := flag.String("func", "", "function file path")
	inputFile := flag.String("input", "", "input file path")

	flag.Parse()

	priv, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", *port))
	ctx, cancel := context.WithCancel(context.Background())
	host, _ := libp2p.New(
		ctx,
		libp2p.ListenAddrs(listen),
		libp2p.Identity(priv),
	)
	log.Infof("Host ID: %s, address: %s", host.ID().Pretty(), host.Addrs())

	if *peers != "" {
		mAddrList := strings.Split(*peers, ",")
		log.Infof("Connecting to %s", mAddrList)

		for _, mAddr := range mAddrList {
			addr, _ := ma.NewMultiaddr(mAddr)
			addrInfo, _ := peer.AddrInfoFromP2pAddr(addr)
			if err := host.Connect(ctx, *addrInfo); err != nil {
				panic(err)
			}
		}
	}

	node := ipec.NewNodeWithHost(ctx, host)

	// task owner
	if *funcFile != "" && *inputFile != "" {
		task, err := node.CreateTask(ctx, *funcFile, *inputFile, "test task")
		if err != nil {
			panic(err)
		}
		node.Dispatch(ctx, task)
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Kill, os.Interrupt)

	select {
	case <-sigChan:
		log.Info("Stopping IPEC process...")
		cancel()
	}
}
