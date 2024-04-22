package main

import (
	"database/sql"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	_ "github.com/lib/pq"
	"log"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgre"
	password = "postgre"
	dbname   = "postgre"
)

type Cluster struct {
	ID      int
	Name    string
	Members []string
}
type ClusterInfo struct {
	ClusterName    string   `json:"cluster_name"`
	MemberCount    int      `json:"cluster_member_count"`
	Page           int      `json:"cluster_page"`
	ClusterMembers []string `json:"cluster_members"`
}
type AddressCluster map[string]int

var clusters []Cluster
var addressToCluster AddressCluster

func main() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to db!")

	connConfig := &rpcclient.ConnConfig{
		Host:         "100.77.25.60:8332",
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

	//Get the current block height
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Total number of blocks:", blockCount)

	addressToCluster = make(AddressCluster)

	//Retrieving data till 3000
	const maxBlockHeight = 3000

	for blockHeight := int64(0); blockHeight <= maxBlockHeight; blockHeight++ {
		blockHash, err := client.GetBlockHash(blockHeight)
		if err != nil {
			log.Fatal(err)
		}

		block, err := client.GetBlock(blockHash)
		if err != nil {
			log.Fatal(err)
		}

		for _, tx := range block.Transactions {
			processTransaction(client, tx)
		}
	}

	storeClustersInDB(db)

	walletAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	clusterInfo := getClusterInfo(db, walletAddress)
	fmt.Printf("Cluster Info for Address %s:\n", walletAddress)
	fmt.Printf("Cluster Name : %s\n", clusterInfo.ClusterName)
	fmt.Printf("Member Count: %d\n", clusterInfo.MemberCount)
	fmt.Printf("Page: %d\n", clusterInfo.Page)
	fmt.Printf("Cluster Members: %v\n", clusterInfo.ClusterMembers)
}

func storeClustersInDB(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS clusters (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    member_count INTEGER NOT NULL,
)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cluster_members (
    id SERIAL PRIMARY KEY,
    cluster_id INTEGER REFERENCES clusters(id),
    address TEXT NOT NULL
)`)
	if err != nil {
		log.Fatal(err)
	}

	for _, cluster := range clusters {
		if len(cluster.Members) == 0 {
			continue
		}

		var clusterID int
		err := db.QueryRow(`INSERT INTO clusters (name, member_count) VALUES ($1, $2) RETURNING id`,
			cluster.Name, len(cluster.Members)).Scan(&clusterID)
		if err != nil {
			log.Fatal(err)
		}
		for _, address := range cluster.Members {
			_, err := db.Exec(`INSERT INTO cluster_members (cluster_id, address) VALUES ($1, $2)`,
				clusterID, address)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func getClusterInfo(db *sql.DB, walletAddress string) ClusterInfo {
	var clusterInfo ClusterInfo

	// Query the cluster information for the given wallet address
	var clusterID int
	var clusterName string
	var memberCount int
	err := db.QueryRow(`SELECT c.id, c.name, c.member_count 
        FROM clusters c
        INNER JOIN cluster_members cm ON c.id = cm.cluster_id
        WHERE cm.address = $1
        LIMIT 1`, walletAddress).Scan(&clusterID, &clusterName, &memberCount)
	if err != nil {
		if err == sql.ErrNoRows {
			// Wallet address not found in any cluster
			return clusterInfo
		}
		log.Fatal(err)
	}

	// Query the cluster members
	rows, err := db.Query(`SELECT address FROM cluster_members WHERE cluster_id = $1 LIMIT 100`, clusterID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var clusterMembers []string
	for rows.Next() {
		var address string
		err := rows.Scan(&address)
		if err != nil {
			log.Fatal(err)
		}
		clusterMembers = append(clusterMembers, address)
	}

	clusterInfo.ClusterName = clusterName
	clusterInfo.MemberCount = memberCount
	clusterInfo.Page = 1
	clusterInfo.ClusterMembers = clusterMembers

	return clusterInfo
}

func processTransaction(client *rpcclient.Client, tx *wire.MsgTx) {
	inputAddresses := extractInputAddresses(client, tx)
	outputAddresses := extractOutputAddresses(tx)

	clusterID := mergeInputAddresses(inputAddresses)

	changeAddress := identifyChangeAddress(tx, outputAddresses)
	if changeAddress != "" {
		mergeAddressToCluster(changeAddress, clusterID)
	}
	for _, addr := range outputAddresses {
		mergeAddressToCluster(addr, clusterID)
	}
}

func extractInputAddresses(client *rpcclient.Client, tx *wire.MsgTx) []string {
	var addresses []string
	for _, txIn := range tx.TxIn {
		prevTxHash := txIn.PreviousOutPoint.Hash
		prevTxOut := txIn.PreviousOutPoint.Index
		prevTx, err := client.GetRawTransaction(&prevTxHash)
		if err != nil {
			log.Printf("Failed to retreive previous transaction: %v", err)
			continue
		}
		prevTxOutData := prevTx.MsgTx().TxOut[prevTxOut]
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(prevTxOutData.PkScript, &chaincfg.MainNetParams)
		if err != nil {
			log.Printf("Failed to extract addresses from script : %v", err)
			continue
		}
		for _, addr := range addrs {
			addresses = append(addresses, addr.EncodeAddress())
		}
	}
	return addresses
}

func extractOutputAddresses(tx *wire.MsgTx) []string {
	var addresses []string
	for _, txOut := range tx.TxOut {
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, &chaincfg.MainNetParams)
		if err != nil {
			log.Printf("Failed to extract addresses from script: %v", err)
			continue
		}
		for _, addr := range addrs {
			addresses = append(addresses, addr.EncodeAddress())
		}
	}
	return addresses
}

func mergeInputAddresses(addresses []string) int {
	clusterId := -1
	for _, addr := range addresses {
		if id, ok := addressToCluster[addr]; ok {
			if clusterId == -1 {
				clusterId = id
			} else if clusterId != id {
				mergeClusters(clusterId, id)
			}
		}
	}
	if clusterId == -1 {
		clusterId = createNewCluster()
	}
	for _, addr := range addresses {
		addressToCluster[addr] = clusterId
	}
	return clusterId
}

func identifyChangeAddress(tx *wire.MsgTx, outputAddresses []string) string {
	return ""
}

func mergeAddressToCluster(address string, clusterID int) {
	if id, ok := addressToCluster[address]; ok {
		if id != clusterID {
			mergeClusters(clusterID, id)
		}
	} else {
		addressToCluster[address] = clusterID
		clusters[clusterID].Members = append(clusters[clusterID].Members, address)
	}
}
func createNewCluster() int {
	clusterID := len(clusters)
	clusters = append(clusters, Cluster{
		ID:      clusterID,
		Name:    fmt.Sprintf("Cluster %d", clusterID),
		Members: []string{},
	})

	return clusterID
}

func mergeClusters(clusterID1, clusterID2 int) {
	if clusterID1 == clusterID2 {
		return
	}
	for _, addr := range clusters[clusterID2].Members {
		addressToCluster[addr] = clusterID1
	}
	clusters[clusterID1].Members = append(clusters[clusterID1].Members, clusters[clusterID2].Members...)
	clusters[clusterID2].Members = nil
}
