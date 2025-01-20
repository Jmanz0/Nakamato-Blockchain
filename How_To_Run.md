# How to run:

## Part 1:

### Generate a chain with 100 nodes:

1) make sure your in root directory nakamoto-blockhain
2) run 'make proto'
3) run 'make deploy'
4) run 'make run_generic_case'
5) run 'make run_client'
6) Once enough time passes, run 'make stop'. This will collect all outputs of the 5 miners to the local ./output/ directory

### Adjusting block generation speed
1) within ./internal/blockchain/blockchain.go:43, change the number of f's to modify the generation speed. Save.
2) run 'make proto' (if not done already)
3) run 'make deploy'
4) run 'make run_generic_case'
5) run 'make run_client'
6) Once enough time passes, run 'make stop'. This will collect all outputs of the 5 miners to the local ./output/ directory

### Corruption case
1) make sure your in root directory nakamoto-blockhain
2) run 'make proto' (if not done already)
3) run 'make deploy'
4) run 'make run_corrupted_case'
5) run 'make run_client'
6) Once enough time passes, run 'make stop'. This will collect all outputs of the 5 miners to the local ./output/ directory

### Lying miner
1) make sure your in root directory nakamoto-blockhain
2) run 'make proto' (if not done already)
3) run 'make deploy'
4) run 'make run_lying_case'
5) run 'make run_client'
6) Once enough time passes, run 'make stop'. This will collect all outputs of the 5 miners to the local ./output/ directory

### Fork in blockchain
1) make sure your in root directory nakamoto-blockhain
2) run 'make proto' (if not done already)
3) run 'make deploy'
4) run 'make run_fork_case' 
5) run 'make run_client'
6) Once enough time passes, run 'make stop'. This will collect all outputs of the 5 miners to the local ./output/ directory

## Phase 2

### Dynamic Mining Speed
1) within ./internal/blockchain/blockchain.go:37, change the DynamicStart variable to 10. 
2) run 'make proto' (if not done already)
3) run 'make deploy'
4) run 'make run_generic_case'
5) run 'make run_client'
6) Once enough time passes, run 'make stop'. This will collect all outputs of the 5 miners to the local ./output/ directory. There will be timestamps to see the processing speed between blocks.

### Client UI
1) make sure your in root directory nakamoto-blockhain
2) run 'make proto' (if not done already)
3) run 'make deploy'
4) run 'make run_generic_case' (this is part of phase 2, but demonstrated lying miner)
5) run 'make run_ui_client'
    a) Within ./config/keys.json you can find the public keys of initial utxos, use this while running the ui_client.
    b) Save the transaction code to check if the transaction is far enough into the chain (add more transactions to see this change)
6) Once enough time passes, run 'make stop'. This will collect all outputs of the 5 miners to the local ./output/ directory

### Blacklist interrupting miners
1) make sure your in root directory nakamoto-blockhain
2) run 'make proto' (if not done already)
3) run 'make deploy'
4) run 'make run_blacklist_case' 
5) run 'make run_client'
6) Once enough time passes, run 'make stop'. This will collect all outputs of the 5 miners to the local ./output/ directory


# Nakamoto Blockchain Originial Architecture (Outdated)

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

UTXO set Structure: confusion????
- Key value store (transaction ID (txid) and output index + output: bal)?

## Testing/Demo

Demo Features:
- Force Fork
- Force Invalid Block
- Force Invalid Transaction (too low funds, wrong signature)
- Adjust PoW difficulty
- Make Transaction

- Monitor mining speed
- Check Balances of public key / address
- Query Blockchain State From Miner, i.e. 0x743 -> 0x853 -> 0x952

## Logging
- Log chain per block update (log block hash + length of chain)?
- Errors
- Rejection/Acceptance of blocks
- Individual finding of block
- Log received transaction
- UTXOs updates (added or removed) (transaction hash + public key + balance)

