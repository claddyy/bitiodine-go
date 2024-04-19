package main

import (
	"log"

	"github.com/btcsuite/btcd/rpcclient"
)

func main() {
	connConfig := &rpcclient.ConnConfig{
		Host:         "100.77.25.60:8332", // Update with the correct RPC host and port for your Bitcoin Core co
		User:         "hornet",
		Pass:         "hornet",
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connConfig, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	// Get the current block height
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Total number of blocks: ", blockCount)
}
