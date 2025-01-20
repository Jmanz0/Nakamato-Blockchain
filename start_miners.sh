#!/bin/bash
# start_miner.sh

if [ $# -ne 2 ]; then
    echo "Usage: $0 <num_miners> <output_dir>"
    exit 1
fi

num_miners=$1
output_dir=$2

if ! [[ "$num_miners" =~ ^[0-9]+$ ]]; then
    echo "Error: <num_miners> must be a positive integer."
    exit 1
fi

base_ip="122.200.68.26"
username="osgroup17"
start_port=8051
end_port=8070
miner_binary="./miner"
miner_list="miners.txt"
initial_utxos_local_path="./config/initial_utxos.json"
initial_utxos="initial_utxos.json"

# Clear existing miner list
> "$miner_list"

# Fetch internal IPs for all miners
fetch_internal_ips() {
    echo "Fetching internal IPs for all miners..."
    miner_count=0
    for port in $(seq "$start_port" "$end_port"); do
        if [ "$miner_count" -ge "$num_miners" ]; then
            echo "Fetched IPs for $num_miners miners."
            break
        fi
        internal_ip=$(ssh -o StrictHostKeyChecking=no -p "$port" "$username@$base_ip" \
            "ip -4 addr show | grep -oP '(?<=inet\s)10\.1\.\d+\.\d+' | head -n 1")
        if [ -n "$internal_ip" ]; then
            echo "$internal_ip" >> "$miner_list"
            echo "Internal IP for miner: $internal_ip"
            miner_count=$((miner_count + 1))
        else
            echo "Warning: Could not retrieve internal IP for miner."
        fi
    done

    if [ "$miner_count" -lt "$num_miners" ]; then
        echo "Error: Could not fetch enough internal IPs. Found $miner_count, expected $num_miners."
        exit 1
    fi
}

# Start miners with the fetched internal IPs, excluding itself from peers list
start_miners() {
    echo "Starting miners using fetched IPs from $miner_list..."
    internal_ips=($(cat "$miner_list"))
    miner_count=0
    http_port=8080
    grpc_port=50051

    for ip in "${internal_ips[@]}"; do
        if [ "$miner_count" -ge "$num_miners" ]; then
            echo "Started $num_miners miners."
            break
        fi

        # Exclude current miner's IP from peers
        peers=()
        for peer_ip in "${internal_ips[@]}"; do
            if [ "$peer_ip" != "$ip" ]; then
                peers+=("$peer_ip")
            fi
        done
        peer_list=$(echo "${peers[@]}" | tr ' ' ',')

        echo "Starting miner on HTTP port $http_port, gRPC port $grpc_port with peers: $peer_list..."

        # Upload the updated miner binary
        scp -o StrictHostKeyChecking=no "$miner_binary" "$username@$ip:/osdata/osgroup17/"
        if [ $? -ne 0 ]; then
            echo "Failed to upload miner binary to IP $ip. Skipping..."
            continue
        fi

        # Upload initial_utxos.json
        scp -o StrictHostKeyChecking=no "$initial_utxos_local_path" "$username@$ip:/osdata/osgroup17/"
        if [ $? -ne 0 ]; then
            echo "Failed to upload miner binary to IP $ip. Skipping..."
            continue
        fi

        echo "Miner files transferred to IP $ip. Miner not started."
        miner_count=$((miner_count + 1))
    done
}

# Main execution
fetch_internal_ips
start_miners
echo "All miner files transferred successfully. Internal IPs and peer information saved to $miner_list. Miners not started."
