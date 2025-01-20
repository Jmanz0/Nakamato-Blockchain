// Reads the keys files to accumulate keys. It then asks for particular public key for each sender and recipient, fetching the key from file dynamically.
package main

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "math/rand"
    "nakamoto-blockchain/internal/blockchain"
    "nakamoto-blockchain/internal/crypto"
    "nakamoto-blockchain/internal/server"
    "nakamoto-blockchain/logger"
    "nakamoto-blockchain/proto/gen"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

var (
    mu      sync.Mutex
    utxoSet = blockchain.NewUTXOSet()
)

func main() {
    logger.Init()

    if len(os.Args) < 4 {
        logger.ErrorLogger.Fatal("[Client] Usage: go run main.go <initialUXTOS_file> <keys_file> <miner_ip_file>")
    }

    initialUXTOSFile := os.Args[1]
    keysFile := os.Args[2]
    minerIPFile := os.Args[3]

    keyMap, err := readKeyMap(keysFile)
    if err != nil {
        logger.ErrorLogger.Fatal("[Client] Failed to read keys:", err)
    }

    err = populateInitialUTXOSet(initialUXTOSFile)
    if err != nil {
        logger.ErrorLogger.Fatal("[Client] Failed to populate UTXOs:", err)
    }

    minerIPs, err := readMinerIPs(minerIPFile)
    if err != nil {
        logger.ErrorLogger.Fatal("[Client] Failed to read miner IPs:", err)
    }
    if len(minerIPs) == 0 {
        logger.ErrorLogger.Fatal("[Client] No miners available")
    }

    // Simple CLI loop
    for {
        fmt.Println("\n=========== MENU ===========")
        fmt.Println("1) Send new transaction")
        fmt.Println("2) Check transaction status (k=3)")
        fmt.Println("3) Get current balance (local)")
        fmt.Println("4) Quit")
        fmt.Print("Enter choice: ")

        var choice string
        fmt.Scanln(&choice)

        switch choice {
        case "1":
            handleSendTransaction(keyMap, minerIPs)
        case "2":
            handleCheckStatus(minerIPs)
        case "3":
            handleGetBalance()
        case "4":
            fmt.Println("[Client UI] Exiting...")
            return
        default:
            fmt.Println("Invalid choice. Try again.")
        }
    }
}

// --------------------
// MENU HANDLERS

func handleSendTransaction(keyMap map[string]string, minerIPs []string) {
    reader := bufio.NewReader(os.Stdin)

    // Prompt sender
    fmt.Print("Enter sender public key: ")
    senderPubKey, _ := reader.ReadString('\n')
    senderPubKey = strings.TrimSpace(senderPubKey)

    // Prompt recipient
    fmt.Print("Enter recipient public key: ")
    recipientPubKey, _ := reader.ReadString('\n')
    recipientPubKey = strings.TrimSpace(recipientPubKey)

    // Prompt amount
    fmt.Print("Enter amount to send: ")
    amountStr, _ := reader.ReadString('\n')
    amountStr = strings.TrimSpace(amountStr)
    amountToSend, err := strconv.ParseInt(amountStr, 10, 64)
    if err != nil || amountToSend <= 0 {
        fmt.Println("Invalid amount.")
        return
    }

    // Acquire private key
    privateKey, ok := keyMap[senderPubKey]
    if !ok {
        fmt.Println("Sender key not found in keys file.")
        return
    }

    // Build transaction
    mu.Lock()
    defer mu.Unlock()

    senderAddress := crypto.Key2Addr(senderPubKey)
    selectedUTXOs, err := utxoSet.GetUTXOs(senderAddress, amountToSend)
    if err != nil {
        fmt.Printf("Failed to select UTXOs: %v\n", err)
        return
    }

    tx, err := blockchain.NewTransaction(selectedUTXOs, senderPubKey, recipientPubKey, amountToSend)
    if err != nil {
        fmt.Printf("Failed to create transaction: %v\n", err)
        return
    }

    if err := tx.Sign(privateKey); err != nil {
        fmt.Printf("Failed to sign transaction: %v\n", err)
        return
    }

    // Send to a random miner
    minerIP := minerIPs[rand.Intn(len(minerIPs))]
    txProto := server.ConvertTransactionToGrpc(tx)
    if err := sendTransactionToMiner(txProto, minerIP+":50051"); err != nil {
        fmt.Printf("Failed to send transaction to %s: %v\n", minerIP, err)
        return
    }

    // Update local UTXO set
    if err := utxoSet.AddTransaction(*tx); err != nil {
        fmt.Printf("Failed to update UTXOSet: %v\n", err)
        return
    }

    fmt.Printf("Transaction %s sent successfully to miner %s.\n", tx.Hash, minerIP)
}

func handleCheckStatus(minerIPs []string) {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter transaction hash to check: ")
    txHash, _ := reader.ReadString('\n')
    txHash = strings.TrimSpace(txHash)

    // We call the new GetTransactionStatus RPC on one (or more) miners. For demonstration, just pick the first miner or random.
    if len(minerIPs) == 0 {
        fmt.Println("No miners to connect to.")
        return
    }
    miner := minerIPs[0]

    confirmed, err := checkTransactionStatusRPC(txHash, miner+":50051", 3)
    if err != nil {
        fmt.Printf("Error checking status on miner %s: %v\n", miner, err)
        return
    }

    if confirmed {
        fmt.Printf("Transaction %s is confirmed (k=3) on miner %s.\n", txHash, miner)
    } else {
        fmt.Printf("Transaction %s not yet confirmed with k=3 on miner %s.\n", txHash, miner)
    }
}

func handleGetBalance() {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter public key to see current local balance: ")
    pubKey, _ := reader.ReadString('\n')
    pubKey = strings.TrimSpace(pubKey)

    address := crypto.Key2Addr(pubKey)

    mu.Lock()
    defer mu.Unlock()

    utxos := utxoSet.Get(address)
    var balance int64
    for _, ut := range utxos {
        balance += ut.Amount
    }

    fmt.Printf("Local balance for %s: %d\n", pubKey, balance)
}

// --------------------
// NEW RPC HELPER

func checkTransactionStatusRPC(txHash, minerAddr string, k int) (bool, error) {
    conn, err := grpc.Dial(minerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return false, fmt.Errorf("failed to connect to miner at %s: %w", minerAddr, err)
    }
    defer conn.Close()

    client := gen.NewIncomingCommunicatorServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req := &gen.TransactionStatusRequest{
        Hash: txHash,
        K:    int32(k),
    }

    resp, err := client.GetTransactionStatus(ctx, req)
    if err != nil {
        return false, fmt.Errorf("GetTransactionStatus RPC error: %w", err)
    }
    if resp.Error != "" {
        return false, fmt.Errorf(resp.Error)
    }
    return resp.Confirmed, nil
}

// --------------------
// UNCHANGED HELPERS

func sendTransactionToMiner(tx *gen.Transaction, minerIP string) error {
    conn, err := grpc.Dial(minerIP, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return fmt.Errorf("failed to connect to miner at %s: %w", minerIP, err)
    }
    defer conn.Close()

    client := gen.NewIncomingCommunicatorServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err = client.SubmitTransaction(ctx, tx)
    if err != nil {
        return fmt.Errorf("failed to submit transaction: %w", err)
    }
    return nil
}

func readMinerIPs(filename string) ([]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var minerIPs []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        minerIPs = append(minerIPs, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    return minerIPs, nil
}

func readKeyMap(filename string) (map[string]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var keyPairs []struct {
        PublicKey  string `json:"public_key"`
        PrivateKey string `json:"private_key"`
    }
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&keyPairs); err != nil {
        return nil, fmt.Errorf("invalid key file format: %w", err)
    }

    keyMap := make(map[string]string)
    for _, keyPair := range keyPairs {
        keyMap[keyPair.PublicKey] = keyPair.PrivateKey
    }
    return keyMap, nil
}

func populateInitialUTXOSet(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("failed to open UTXO file: %w", err)
    }
    defer file.Close()

    var utxoEntries []struct {
        PublicKey string            `json:"public_key"`
        UTXOs     []blockchain.UTXO `json:"utxos"`
    }
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&utxoEntries); err != nil {
        return fmt.Errorf("invalid UTXO format: %w", err)
    }

    mu.Lock()
    defer mu.Unlock()
    for _, entry := range utxoEntries {
        address := crypto.Key2Addr(entry.PublicKey)
        for _, utxo := range entry.UTXOs {
            utxo.Address = address
            utxoSet.AddUTXO(utxo)
        }
    }
    return nil
}