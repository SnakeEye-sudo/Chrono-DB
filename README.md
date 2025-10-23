# Chrono-DB

A distributed temporal database allowing queries of any historical state with millisecond precision using CRDTs for multi-master replication and Raft consensus for strong consistency.

## 🚀 Features

- **Bitemporal Data Model**: Query data as it was known at any point in time (valid time) and when it was recorded (transaction time)
- **CRDT-based Replication**: Conflict-free multi-master replication for high availability
- **Raft Consensus**: Strong consistency guarantees across distributed nodes
- **Millisecond Precision**: Query historical states with millisecond-level accuracy
- **RESTful API**: Simple HTTP API for all operations
- **CLI Client**: Command-line tool for easy interaction

## 📋 Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Chrono-DB Cluster                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐      ┌──────────┐      ┌──────────┐         │
│  │  Node 1  │◄────►│  Node 2  │◄────►│  Node 3  │         │
│  │  Leader  │      │ Follower │      │ Follower │         │
│  └────┬─────┘      └────┬─────┘      └────┬─────┘         │
│       │                 │                  │                │
│  ┌────▼─────────────────▼──────────────────▼────┐          │
│  │         Raft Consensus Layer                  │          │
│  │  (Leader Election & Log Replication)          │          │
│  └───────────────────┬───────────────────────────┘          │
│                      │                                       │
│  ┌───────────────────▼───────────────────────────┐          │
│  │            CRDT Store                          │          │
│  │  (GCounter, LWW-Register)                      │          │
│  └───────────────────┬───────────────────────────┘          │
│                      │                                       │
│  ┌───────────────────▼───────────────────────────┐          │
│  │         Bitemporal Engine                      │          │
│  │  (Valid Time & Transaction Time)               │          │
│  └────────────────────────────────────────────────┘          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌───────────────┐
                  │   REST API    │
                  │  (Port 8080)  │
                  └───────────────┘
```

## 🛠️ Installation

### Prerequisites

- Go 1.19 or higher
- Git

### Clone the Repository

```bash
git clone https://github.com/SnakeEye-sudo/Chrono-DB.git
cd Chrono-DB
```

### Initialize Go Module

```bash
go mod init github.com/SnakeEye-sudo/Chrono-DB
go mod tidy
```

### Build the Server

```bash
go build -o chrono-db main.go db_engine.go crdt.go raft.go api_server.go
```

### Build the CLI Client

```bash
go build -o chrono-client client.go
```

## 🚦 Quick Start

### Start a Single Node

```bash
./chrono-db -node=node1 -http=8080 -raft=9000 -data=./data
```

### Verify Server is Running

```bash
curl http://localhost:8080/
```

Expected response:
```json
{
  "service": "Chrono-DB",
  "version": "1.0.0",
  "description": "Distributed temporal database with CRDT and Raft consensus"
}
```

## 📡 API Reference

### 1. Insert Data

**Endpoint:** `POST /api/v1/insert`

```bash
curl -X POST http://localhost:8080/api/v1/insert \
  -H "Content-Type: application/json" \
  -d '{
    "key": "user:1001",
    "value": {
      "name": "Alice Johnson",
      "email": "alice@example.com",
      "balance": 5000
    },
    "valid_start": "2024-01-01T00:00:00Z",
    "valid_end": "9999-12-31T23:59:59Z"
  }'
```

Response:
```json
{
  "status": "success",
  "key": "user:1001"
}
```

### 2. Query Current Value

**Endpoint:** `GET /api/v1/query?key={key}`

```bash
curl "http://localhost:8080/api/v1/query?key=user:1001"
```

Response:
```json
{
  "found": true,
  "key": "user:1001",
  "value": {
    "name": "Alice Johnson",
    "email": "alice@example.com",
    "balance": 5000
  }
}
```

### 3. Temporal Query (Point-in-Time)

**Endpoint:** `GET /api/v1/temporal?key={key}&as_of={timestamp}&valid_time={timestamp}`

```bash
curl "http://localhost:8080/api/v1/temporal?key=product:SKU-001&valid_time=2024-03-15T00:00:00Z"
```

Response:
```json
{
  "found": true,
  "key": "product:SKU-001",
  "value": {
    "name": "Laptop Pro 15",
    "price": 1299.99
  },
  "as_of": "2024-10-23T14:30:00Z",
  "valid_time": "2024-03-15T00:00:00Z"
}
```

### 4. Get History

**Endpoint:** `GET /api/v1/history?key={key}`

```bash
curl "http://localhost:8080/api/v1/history?key=product:SKU-001"
```

Response:
```json
{
  "key": "product:SKU-001",
  "history": [
    {
      "key": "product:SKU-001",
      "value": {"price": 1299.99},
      "valid_time_start": "2024-02-01T00:00:00Z",
      "valid_time_end": "2024-06-30T23:59:59Z",
      "transaction_time": "2024-02-01T10:30:00Z"
    },
    {
      "key": "product:SKU-001",
      "value": {"price": 1199.99},
      "valid_time_start": "2024-07-01T00:00:00Z",
      "valid_time_end": "9999-12-31T23:59:59Z",
      "transaction_time": "2024-07-01T08:00:00Z"
    }
  ]
}
```

### 5. Cluster Status

**Endpoint:** `GET /api/v1/status`

```bash
curl http://localhost:8080/api/v1/status
```

Response:
```json
{
  "node_id": "node1",
  "raft_state": "leader",
  "raft_term": 1,
  "timestamp": "2024-10-23T14:30:00Z"
}
```

### 6. CRDT Counter Operations

**Increment Counter:**
```bash
curl -X POST "http://localhost:8080/api/v1/crdt/counter?key=page_views"
```

**Get Counter Value:**
```bash
curl "http://localhost:8080/api/v1/crdt/counter?key=page_views"
```

## 🖥️ CLI Client Usage

### Insert Data

```bash
./chrono-client insert user:1002 '{"name":"Bob Smith","email":"bob@example.com"}'
```

### Query Data

```bash
./chrono-client query user:1002
```

### Get History

```bash
./chrono-client history product:SKU-001
```

### Check Status

```bash
./chrono-client status
```

### Using Custom API URL

```bash
./chrono-client -url=http://localhost:8081 query user:1001
```

## 🏗️ Running a Multi-Node Cluster

### Start Node 1 (Leader)

```bash
./chrono-db -node=node1 -http=8080 -raft=9000 -data=./data/node1
```

### Start Node 2

```bash
./chrono-db -node=node2 -http=8081 -raft=9001 -data=./data/node2 -join=localhost:9000
```

### Start Node 3

```bash
./chrono-db -node=node3 -http=8082 -raft=9002 -data=./data/node3 -join=localhost:9000
```

Each node will:
- Sync with the leader
- Participate in consensus
- Replicate data via CRDT

## 🧪 Testing

Load test data from `testdata.json`:

```bash
# Insert test records
curl -X POST http://localhost:8080/api/v1/insert \
  -H "Content-Type: application/json" \
  -d @testdata.json
```

Run API tests:

```bash
# Test 1: Query current user data
curl "http://localhost:8080/api/v1/query?key=user:1001"

# Test 2: Query historical product price
curl "http://localhost:8080/api/v1/temporal?key=product:SKU-001&valid_time=2024-03-15T00:00:00Z"

# Test 3: Get full history
curl "http://localhost:8080/api/v1/history?key=product:SKU-001"
```

## 📊 Use Cases

### 1. Financial Systems
- Track account balances over time
- Audit trail for all transactions
- Regulatory compliance with historical queries

### 2. E-commerce
- Price history tracking
- Inventory snapshots at any point in time
- Order status evolution

### 3. Configuration Management
- Track configuration changes
- Rollback to any previous state
- Audit who changed what and when

### 4. Healthcare Records
- Patient history tracking
- Treatment timeline
- Compliance with data retention policies

## 🔧 Configuration

See `example_config.json` for cluster configuration options:

- **Node settings**: ID, ports, data directory
- **Raft parameters**: Election timeout, heartbeat interval
- **Database options**: History retention, compaction
- **API limits**: Timeouts, request size limits

## 🛡️ Technical Details

### Bitemporal Model

- **Valid Time**: When the fact was true in reality
- **Transaction Time**: When the fact was recorded in the database

This allows queries like:
- "What did we know about X on date Y?"
- "What was the actual value of X on date Y?"
- "Show me all changes to X"

### CRDT Implementation

- **GCounter**: Grow-only counter for distributed counting
- **LWW-Register**: Last-Writer-Wins register for conflict resolution

### Raft Consensus

- Leader election with randomized timeouts
- Log replication with commit index tracking
- Automatic failover on leader failure

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - feel free to use this project for learning and development.

## 🙏 Acknowledgments

- Inspired by bitemporal database concepts
- CRDT research from various academic papers
- Raft consensus algorithm by Diego Ongaro and John Ousterhout

---

**Built with ❤️ for distributed systems enthusiasts**
