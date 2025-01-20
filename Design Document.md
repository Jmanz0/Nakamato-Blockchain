# Design Document

## System Components

### 1. UTXO (Unspent Transaction Output)

- **Structure**:
  - Transaction ID (SHA256 hash)
  - Index
  - Amount
  - Address

### 2. Transaction

- **Components**:
  - Input UTXOs (list)
  - Output UTXOs (list)
  - Sender's Public Key
  - Transaction Hash (SHA256)
  - Digital Signature (ECDSA)
- **Validation**:
  - Verify signature using sender's public key
  - Ensure input UTXOs exist and are unspent
  - Verify input amount equals output amount

### 3. Block

- **Header**:
  - Timestamp
  - Previous Block Hash
  - Content Hash
  - Height
  - Difficulty
  - Nonce
- **Content**:
  - List of transactions
- **Additional**:
  - Block Hash (SHA256)

### 4. Blockchain

- **Components**:
  - Block list
  - Unused UTXO set
- **Maintenance**:
  - Track and update UTXO set with each new block
  - Maintain chain consistency

## Network Architecture

### Miner Node

- **RPC Server Functions**:
  1. `GetBlock(hash)`: Gets block by its hash
  2. `SubmitTransaction(tx)`: Submit new transaction
  3. `SubmitBlock(block, headers)`: Submit new block
  4. `GetChain()`: Retrieve blockchain data

### Client

- **Functionality**:
  - Submit transactions via miner's RPC
  - Maintain local chain copy
  - Calculate balance from local chain

## Implementation

### Client

1. Pick out unused UTXOs from the UTXO Pool to form a transaction, and signs it. (Currently client keeps track of its own UTXOs)
2. Submit the transaction to one miner with `SubmitTransaction()`.
3. TODO: Periodically calls `GetChain` from a miner to update its local blockchain.

### Miner

1. When receiving a block via `SubmitTransaction`, first verifies the hash and signature, then broadcasts it to all other miners by calling `SubmitTransaction` of other miners and adds it to the transaction pool.
2. Selects a batch of transactions that uses valid unused UTXOs, and creates a new block with them.
3. The difficulty of the new block is determined dynamically, by checking the timestamps of block `n-1` and block `n-10`. The miner can calculate the average mining time of the last 10 blocks, and calculates the new difficulty based on them.
4. Tries to find a valid nonce of the block, making the hash less than the difficulty.
5. After finding a valid nonce, submits this block, with the previous 100 block headers by calling `SubmitBlock` of other miners.
6. When receiving a new block with previous 100 block headers, the miner first verifies the hash of the block headers are correct and matches the difficulty. Then, it checks if the received chain correctly connects to the existing chain. If so, and the new chain is longer than the old chain, it updates the local chain to the new chain. If it finds valid blocks in the submitted headers, but it don't have the block locally, it calls `GetBlock` to get the block by the hash.

## Improvement Plans

1. Use Merkle tree instead of using one hash for all transactions in a block.
2. Make miners capable of fetching missing blocks before the block when a valid block that doesn't connect to the chain is submitted, thus removing the need to send previous 100 block headers along with the new block.
3. Use a better broadcast protocol that doesn't just broadcast to a set list of miners, and implementing dynamically adding/removing miners.