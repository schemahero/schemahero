.PHONY: all postgres mysql timescaledb sqlite rqlite cassandra clean test test-all test-postgres test-mysql test-timescaledb test-sqlite test-rqlite test-cassandra

OUTPUT_DIR ?= ./bin

all: postgres mysql timescaledb sqlite rqlite cassandra

postgres:
	@echo "Building postgres plugin..."
	@mkdir -p $(OUTPUT_DIR)
	cd postgres && go build -o ../$(OUTPUT_DIR)/schemahero-postgres .
	@echo "Built: $(OUTPUT_DIR)/schemahero-postgres"

mysql:
	@echo "Building mysql plugin..."
	@mkdir -p $(OUTPUT_DIR)
	cd mysql && go build -o ../$(OUTPUT_DIR)/schemahero-mysql .
	@echo "Built: $(OUTPUT_DIR)/schemahero-mysql"

timescaledb:
	@echo "Building timescaledb plugin..."
	@mkdir -p $(OUTPUT_DIR)
	cd timescaledb && go build -o ../$(OUTPUT_DIR)/schemahero-timescaledb .
	@echo "Built: $(OUTPUT_DIR)/schemahero-timescaledb"

sqlite:
	@echo "Building sqlite plugin..."
	@mkdir -p $(OUTPUT_DIR)
	cd sqlite && go build -o ../$(OUTPUT_DIR)/schemahero-sqlite .
	@echo "Built: $(OUTPUT_DIR)/schemahero-sqlite"

rqlite:
	@echo "Building rqlite plugin..."
	@mkdir -p $(OUTPUT_DIR)
	cd rqlite && go build -o ../$(OUTPUT_DIR)/schemahero-rqlite .
	@echo "Built: $(OUTPUT_DIR)/schemahero-rqlite"

cassandra:
	@echo "Building cassandra plugin..."
	@mkdir -p $(OUTPUT_DIR)
	cd cassandra && go build -o ../$(OUTPUT_DIR)/schemahero-cassandra .
	@echo "Built: $(OUTPUT_DIR)/schemahero-cassandra"

clean:
	@echo "Cleaning plugin binaries..."
	@rm -rf $(OUTPUT_DIR)
	@echo "Clean complete"

test: test-postgres test-mysql test-timescaledb test-sqlite test-rqlite test-cassandra

test-all:
	@echo "Testing all plugins..."
	go test -v ./...

test-postgres: postgres
	@echo "Testing postgres plugin..."
	cd postgres && go test -v ./...

test-mysql: mysql
	@echo "Testing mysql plugin..."
	cd mysql && go test -v ./...

test-timescaledb: timescaledb
	@echo "Testing timescaledb plugin..."
	cd timescaledb && go test -v ./...

test-sqlite: sqlite
	@echo "Testing sqlite plugin..."
	cd sqlite && go test -v ./...

test-rqlite: rqlite
	@echo "Testing rqlite plugin..."
	cd rqlite && go test -v ./...

test-cassandra: cassandra
	@echo "Testing cassandra plugin..."
	cd cassandra && go test -v ./...