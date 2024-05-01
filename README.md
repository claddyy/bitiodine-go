# Bitcoin Clustering
This project demonstrates how to extract intelligence from the Bitcoin network by performing co-spend clustering on Bitcoin addresses. It reads the "BitIodine: Extracting Intelligence from the Bitcoin Network" research paper and implements the clustering heuristics described in the paper.

Prerequisites

```Go programming language
PostgreSQL database
Bitcoin Core node
```
Installation

Clone the repository:
git clone https://github.com/your-username/bitcoin-clustering.git

Install the required dependencies:
go mod download

Set up the PostgreSQL database:

Create a new database named postgre.
Update the database connection details in the main function of main.go.


Configure the Bitcoin Core node:

Ensure that your Bitcoin Core node is running and has RPC access enabled.
Update the connConfig in the main function of main.go with the correct RPC host, port, username, and password.



Usage

Run the program:
`go run main.go`

The program will connect to the Bitcoin node, retrieve blocks up to block height 3000, and perform co-spend clustering on the addresses found in the transactions. The cluster information will be stored in the PostgreSQL database.
Retrieving Cluster Information:

The program demonstrates how to retrieve cluster information for a specific wallet address.
Modify the walletAddress variable in the main function to specify the desired wallet address.
The program will query the database and display the cluster information for the specified wallet address.



Database Schema
The project uses the following database schema:

clusters table:

id: Serial primary key
name: Text field representing the cluster name
member_count: Integer field representing the number of members in the cluster


cluster_members table:

id: Serial primary key
cluster_id: Integer field referencing the id of the associated cluster
address: Text field representing the Bitcoin address



### Functions
The project includes the following main functions:

`main`: The entry point of the program. It establishes connections to the PostgreSQL database and Bitcoin node, retrieves blocks up to block height 3000, performs co-spend clustering, stores the clusters in the database, and retrieves cluster information for a specific wallet address.

`storeClustersInDB`: Stores the clusters and their members in the PostgreSQL database.

`getClusterInfo`: Retrieves the cluster information for a given wallet address from the database.

`processTransaction`: Processes a single transaction, extracting input and output addresses, and performs co-spend clustering.

`extractInputAddresses`: Extracts the input addresses from a transaction.

`extractOutputAddresses`: Extracts the output addresses from a transaction.

`mergeInputAddresses`: Merges the input addresses into clusters.

`identifyChangeAddress`: Identifies the change address in a transaction (not implemented in this code).

`mergeAddressToCluster`: Merges an address into a cluster.
createNewCluster: Creates a new cluster.

`mergeClusters`: Merges two clusters together.