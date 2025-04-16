# Custom Ingestion Worker

A Go-based ingestion service that processes records from a JSON file, applies validation and rate limiting, logs failures, and **inserts valid records into MongoDB**.

---

## 📘 Overview

This worker reads records from `samples.json`, validates each, applies per-client rate limiting, and:

- **Inserts valid records into MongoDB**
- **Logs invalid or rejected records** into `error.log`

Built for scalable, testable ingestion flows.

---

## ⚙️ Setup & Run

1. **Clone the repo**

   ```bash
   git clone https://github.com/err-him/custom-ingestion-worker.git
   cd custom-ingestion-worker
   ```

2. **Install dependencies**

   ```bash
   go mod tidy
   ```

3. **Run the worker**

   ```bash
   go run main.go
   ```

---

## 🦪 Input Format (`samples.json`)

```json
[
  {
    "id": "record-001",
    "client_id": "client-A",
    "email": "user@example.com"
  },
  {
    "id": "record-002",
    "client_id": "client-A",
    "email": "bad-email"
  }
]
```

---

## ✅ Success Path

If a record:

- Has a valid email
- Has not exceeded the client’s rate limit

Then:

- It is inserted into MongoDB (`InsertOne`)
- A success message is printed

Example MongoDB document:

```json
{
  "_id": ObjectId("..."),
  "id": "record-001",
  "client_id": "client-A",
  "email": "user@example.com"
}
```

---

## ⚠️ Error Handling

| Type                 | Trigger                                  | Outcome                         |
|----------------------|-------------------------------------------|----------------------------------|
| Invalid Email        | Fails regex/email format                  | Logged to `error.log`            |
| Rate Limit Exceeded  | > 5 records from same `client_id`         | Logged to `error.log`            |
| MongoDB Error        | Insert fails                              | Logged to `error.log`            |
| JSON Parse Error     | Malformed `samples.json`                  | Fatal: program exits             |

### 📄 `error.log` Example

```
2025-04-16T20:35:12+05:30 - Error processing record ID record-002: invalid email format
2025-04-16T20:35:15+05:30 - Error processing record ID record-006: rate limit exceeded for client: client-A
```

---

## 🚦 Rate Limiting

- Each `client_id` is limited to **5 successful inserts**
- Controlled using an in-memory map:

```go
if p.ClientCounts[record.ClientID] >= 5 {
    return fmt.Errorf("rate limit exceeded for client: %s", record.ClientID)
}
```

---

## 📁 Project Structure

```
custom-ingestion-worker/
├── pkg/processor/               # Core business logic
│   ├── processor.go
│   └── processor_test.go
├── main.go                      # Entrypoint
├── samples.json                 # Input records
├── error.log                    # Logs failed records
├── .env                         # MongoDB config
├── go.mod / go.sum              # Dependencies
└── README.md                    # You're here
```

---

## 🧪 Run Tests

```bash
go test ./...
```

---