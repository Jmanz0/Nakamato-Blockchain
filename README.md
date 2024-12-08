# Nakamato-Blockchain
Designed and implemented a blockchain system adhering to the Nakamoto (Bitcoin) protocol, focusing on distributed systems and cryptographic principles. The project includes a client-miner architecture to enable secure transactions, decentralized consensus, and efficient communication between peers.

Features
- Client: RPC-based program to submit transactions to miners.
- Miner: Validates transactions, resolves forks, applies the longest chain rule, and mines blocks with adjustable Proof-of-Work difficulty.
- Networking: Static peer list for transaction and block broadcasting.

# Nakamoto Blockchain

Requirements:
- UTXO Model
- No Merkle Tree


Miner:
- Store Blockchain
- UTXO set
- unprocessed transaction pool

- Static ip table of miners

Miner Features:
- Resolve forks
- Validate blocks + chain
- Validate Transactions
- Broadcast blocks
- Disseminate transaction

Client:
- Send money, submit transaction to Miner using RPC

## Structures

Transaction structure:
- Sender public key
- Receiver public key
- Amount
- Transaction Fee
- Timestamp
- Signature -> private key hash of above
- Transaction Hash -> hash of above

Block Structure:
- block number
- timestamp
- Previous hash
- Nonce

- Hashing algo -> hash each individual transactions -> hash together
- Transactions is just a list?
- block hash

Block-Chain Structure:

UTXO set Structure
- Key value store (transaction ID (txid))

## Forking Algorithm:
- Simplified Forking algorithm through requesting past x block headers


(Currently private repository)
