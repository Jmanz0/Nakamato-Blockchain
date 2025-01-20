#!/bin/bash

if [ $# -ne 1 ]; then
    echo "Usage: $0 <file_with_ip>"
    exit 1
fi

file_with_ip=$1

# Ensure the file exists and is readable
if [ ! -f "$file_with_ip" ] || [ ! -r "$file_with_ip" ]; then
    echo "Error: File '$file_with_ip' does not exist or is not readable."
    exit 1
fi

username="osgroup17"
miner_count=0

# Function to stop a server on a specific IP
stop_server() {
    local ip=$1

    echo "Connecting to $ip..."

    # Kill the server process using pkill
    ssh -o StrictHostKeyChecking=no "$username@$ip" "pkill -x miner"
    if [ $? -eq 0 ]; then
        echo "Miner on $ip has been stopped."
    else
        echo "No running miner found on $ip or failed to stop the miner."
    fi
}

# Read the file line by line
while IFS= read -r ip; do
    if [[ -z "$ip" ]]; then
        echo "Skipping invalid line: $ip"
        continue
    fi

    stop_server "$ip" &
    miner_count=$((miner_count + 1))
done < "$file_with_ip"

# Wait for all background processes to finish
wait

echo "A total of $miner_count miners have been stopped."