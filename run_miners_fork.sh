#!/bin/bash
# run_miners.sh

if [ $# -ne 1 ]; then
    echo "Usage: $0 <output_dir>"
    exit 1
fi

output_dir=$1

username="osgroup17"
http_port=8080
grpc_port=50051
initial_utxos="initial_utxos.json"

# Read internal IPs from the miner list
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

    echo "Starting miner on IP $ip with HTTP port $http_port, gRPC port $grpc_port, and peers: $peer_list..."
    # Start the miner remotely

    echo "ssh -o StrictHostKeyChecking=no $username@$ip"
    echo "rm -rf $output_dir && mkdir -p $output_dir"
    echo "nohup ./miner $initial_utxos $http_port $grpc_port 0 $peer_list > $output_dir/miner_$miner_count.log 2>&1 &"

    ssh -o StrictHostKeyChecking=no "$username@$ip" "
        cd /osdata/osgroup17 &&
        rm -rf $output_dir && mkdir -p $output_dir &&
        screen -dmS miner_session bash -c './miner $initial_utxos $http_port $grpc_port 3 $peer_list > $output_dir/miner_$miner_count.log 2>&1'
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