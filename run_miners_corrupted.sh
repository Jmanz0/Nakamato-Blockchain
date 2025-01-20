#!/bin/bash
# run_miners_corrupted.sh

if [ $# -ne 2 ]; then
    echo "Usage: $0 <output_dir> <second_miner_mode>"
    exit 1
fi

output_dir=$1
corrupted_miner_mode=$2

username="osgroup17"
http_port=8080
grpc_port=50051
initial_utxos="initial_utxos.json"

# Read first 2 internal IPs from the miner list
internal_ips=($(cat miners.txt))

miner_count=0
for ip in "${internal_ips[@]}"; do
    # Exclude the current miner's IP from peers
    peers=()
    for peer_ip in "${internal_ips[@]}"; do
        if [ "$peer_ip" != "$ip" ]; then
            peers+=("$peer_ip")
        fi
    done
    peer_list=$(echo "${peers[@]}" | tr ' ' ',')

    # Set mode - first miner always 0, second uses provided mode
    if [ $miner_count -eq 0 ]; then
        mode=$corrupted_miner_mode
    else
        mode=0
    fi
    
    echo "Starting miner on IP $ip with HTTP port $http_port, gRPC port $grpc_port, mode $mode, and peers: $peer_list..."
    # Start the miner remotely

    echo "ssh -o StrictHostKeyChecking=no $username@$ip"
    echo "rm -rf $output_dir && mkdir -p $output_dir"
    echo "nohup ./miner $initial_utxos $http_port $grpc_port $mode $peer_list > $output_dir/miner_$miner_count.log 2>&1 &"

    ssh -o StrictHostKeyChecking=no "$username@$ip" "
        cd /osdata/osgroup17 &&
        rm -rf $output_dir && mkdir -p $output_dir &&
        screen -dmS miner_session bash -c './miner $initial_utxos $http_port $grpc_port $mode $peer_list > $output_dir/miner_$miner_count.log 2>&1'
        exit
    "
    if [ $? -eq 0 ]; then
        echo "Miner started successfully on IP $ip."
    else
        echo "Failed to start miner on IP $ip."
    fi

    miner_count=$((miner_count + 1))
done

echo "All miners started."
