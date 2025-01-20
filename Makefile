# Variables
MINER_NAME = miner
CLIENT_NAME = client
MINER_FILE = ./cmd/miner/main.go
CLIENT_FILE = ./cmd/client/main.go
UI_CLIENT_FILE = ./cmd/clientui/main.go
PROTO_FILE = proto/blockchain.proto
IP_FILE = miners.txt
REMOTE_OUTPUT_DIR = /osdata/osgroup17/output/
LOCAL_OUTPUT_DIR = output/

all: build

build:
	@echo "Building $(MINER_NAME) and $(CLIENT_NAME)..."
	go build -o $(MINER_NAME) $(MINER_FILE)
	go build -o $(CLIENT_NAME) $(CLIENT_FILE)

run_initial_config:
	go run ./config/key/key_generator.go -count 10 -output ./config/
	go run ./config/utxo/initial_utxo_generator.go -keys ./config/keys.json -output ./config/

upload_miners:
	@echo "Running server with 5 miners server..."
	./start_miners.sh 5 output/

stop_miners:
	./stop_miners.sh $(IP_FILE)

get_miner_stats:
	@echo "Downloading results from miners..."
	mkdir -p $(LOCAL_OUTPUT_DIR)
	@while IFS= read -r miner; do \
		scp -o StrictHostKeyChecking=no -r $$miner:$(REMOTE_OUTPUT_DIR) $(LOCAL_OUTPUT_DIR); \
	done < $(IP_FILE)

run_client:
	@echo "Running client..."
	go run $(CLIENT_FILE) config/initial_utxos.json config/keys.json miners.txt

run_ui_client:
	@echo "Running client..."
	go run $(UI_CLIENT_FILE) config/initial_utxos.json config/keys.json miners.txt

run_generic_case: 
	./run_miners.sh $(REMOTE_OUTPUT_DIR)

run_corrupted_case: 
	./run_miners_corrupted.sh $(REMOTE_OUTPUT_DIR) 1

run_lying_case: 
	./run_miners_corrupted.sh $(REMOTE_OUTPUT_DIR) 2

run_blacklist_case:
	./run_miners_blacklist.sh $(REMOTE_OUTPUT_DIR)

run_fork_case:
	./run_miners_fork.sh $(REMOTE_OUTPUT_DIR)

deploy: build upload_miners

stop: stop_miners get_miner_stats

clean:
	@echo "Cleaning up..."
	rm -f $(MINER_NAME)
	rm -rf proto/*.pb.go

proto: $(PROTO_FILE)
	@echo "Compiling protobuf file..."
	protoc --go_out=. --go-grpc_out=. $^


test:
	@echo "Running tests..."
	go test ./... -v


tidy:
	@echo "Tidying up modules..."
	go mod tidy