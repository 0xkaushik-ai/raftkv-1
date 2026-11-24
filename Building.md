# Milestone 1 
- Start 3 server that talk => Every second each one send a pinf to the other and prints  what happens . If you kill one server the other print that they cant reach it .

- This milestone covers: 
  - Goroutine: doing several things at once 
  - RPC calling a function on another server over the network 
  - Timers = Doing something every second 
  - Timeouts = Not getting stuck waiting on dead server 


- Step 1 
  - cd Raft 
  - mkdir raftkv 
  - cd raftkv 
  - go mod init raftkv 
  - mkdir raft 
  - touch main.go raft/node.go raft/rpc.go
# Learning from milestone 1 
- We built 3 sepearate programs that talk to each other over the network ,keep working when one diess and recover when it comes back . That the foundation of every distributed system .
- Here we ran multiple servers with the same program that run 3 time with the different setting (-id,-addr,-peers) and each one independent with it own memory and its own process .


## One code, three servers
- We wrote the code **once** and ran it **three times**. Each run became a **separate server**.

```
         Same code (raftkv)
                │
    ┌───────────┼───────────┐
    ▼           ▼           ▼
 go run      go run      go run
 -id 1       -id 2       -id 3
 :5001       :5002       :5003
    │           │           │
    ▼           ▼           ▼
 Server 1    Server 2    Server 3
(process)   (process)   (process)
```

- Each `go run .` makes the operating system start a **new process**: a fresh copy of the program with its **own memory**.
- The three copies:
  - run the **same code**
  - get **different flags** (id, port, peers)
  - **don't share memory**, so Server 1 can't see Server 2's variables
- The only way they communicate is **over the network** through RPC.

### Simple analogy
- It's like a **printed instruction manual** given to 3 different people:
  - All three follow the **same instructions** (the code).
  - Each has a different **name tag and desk number** (`-id`, `-addr`).
  - Each has their **own notebook** (memory), and nobody can read anyone else's.
  - To share anything, they have to **call each other on the phone** (RPC).

### Why the ports must differ
- All 3 run on the same laptop (`localhost`), so each needs its **own port**:
  - Server 1 → `5001`
  - Server 2 → `5002`
  - Server 3 → `5003`
- If two servers used the same port, the second would fail with *"address already in use"*. Only one program can listen on a port at a time.

### Real systems work the same way
- In a real deployment, you'd run the **same program on 3 different machines**:
```
Machine A (10.0.0.1):  raftkv -id 1 -addr 10.0.0.1:5000 ...
Machine B (10.0.0.2):  raftkv -id 2 -addr 10.0.0.2:5000 ...
Machine C (10.0.0.3):  raftkv -id 3 -addr 10.0.0.3:5000 ...
```
- The code doesn't change at all, only the addresses. Running 3 processes on one laptop is just a convenient way to **simulate** 3 machines while learning.

**Same code, different settings, separate processes: that's how one program becomes a cluster.**

# Networking concepts used in Milestone 1

## 1. IP address and `localhost`
- **What it is:** an address that identifies a computer on a network. `localhost` (127.0.0.1) means **"this same computer."**
- **Where:** the `-addr localhost:5001` flag.
- **Why:** all 3 servers run on your laptop, so they reach each other through `localhost`.

## 2. Port
- **What it is:** a number (0–65535) that identifies **which program** on a computer gets a message. If the IP address is the building, the port is the flat number.
- **Where:** `5001`, `5002`, `5003`.
- **Why:** 3 programs on the same computer each need their own "door" so messages reach the right one.

## 3. TCP (Transmission Control Protocol)
- **What it is:** a reliable way to send data between programs. It guarantees that data **arrives, in order, without corruption**, or tells you the connection failed.
- **Where:** `"tcp"` in `net.Listen("tcp", ...)` and `net.DialTimeout("tcp", ...)`.
- **Why:** RPC messages must arrive complete and in order.

## 4. Client and server
- **What it is:**
  - **Server:** waits for others to connect.
  - **Client:** starts the connection.
- **Where:** each Node is **both**. It's a server when it receives pings (`acceptLoop`) and a client when it sends them (`sendPing`).
- **Why:** in Raft, every server both sends messages and receives them.

## 5. Listen and accept (the server side)
- **What it is:**
  - **Listen:** "open my port and wait for connections."
  - **Accept:** "someone knocked, let them in." This gives you a **connection**.
- **Where:** `node.go` uses `net.Listen` in `Start()` and `n.listener.Accept()` in `acceptLoop()`.
- **Analogy:** a shop opens its door (listen), and each customer who walks in gets served (accept).

## 6. Dial (the client side)
- **What it is:** connecting **to** another program's IP and port.
- **Where:** `net.DialTimeout("tcp", peerAddr, ...)` in `call()`.
- **Analogy:** dialing someone's phone number.

## 7. Connection
- **What it is:** an open two-way channel between two programs, created by **dial + accept**. Both sides can send and receive on it.
- **Where:** `conn` in `acceptLoop` (server side) and in `call` (client side). `client.Close()` / `conn.Close()` hangs up.
- **Why close it:** every open connection uses resources. Forgetting to close them is a common bug (a connection leak).

## 8. RPC (Remote Procedure Call)
- **What it is:** calling a **function on another computer** as if it were local. The RPC library handles the details:
  1. It converts `PingArgs` into bytes (**serialization**).
  2. It sends them over the TCP connection.
  3. The other side converts the bytes back and runs `Ping`.
  4. The `PingReply` comes back the same way.
- **Where:**
  - `rpc.NewServer()` + `RegisterName("Raft", n)` publish your methods under the name `"Raft"`.
  - `server.ServeConn(conn)` handles RPC requests on a connection.
  - `rpc.NewClient(conn)` + `client.Go("Raft.Ping", ...)` call a remote method.
  - `PingArgs` / `PingReply` are the **message formats**.
- **Why:** Raft is defined in terms of RPCs (`RequestVote`, `AppendEntries`), so this is the main tool we'll keep using.

## 9. Timeouts and network failures
- **What it is:** on a network, a message can be **lost or delayed**, or the other side can be **dead**. A timeout means "if nothing comes back in X time, give up."
- **Where:**
  - `DialTimeout(..., 5*time.Second)`: give up if the server can't be reached.
  - `select` with `time.After(...)` in `call()`: give up if the reply never arrives.
- **Two kinds of failure we saw:**

| Situation | What happens |
|---|---|
| Server killed (port closed) | Dial fails **immediately** with "connection refused" |
| Server frozen or network slow | Nothing comes back, so the **timeout** fires |

- **Key lesson:** you **can't tell "dead" from "slow"** over a network. Raft is designed around this problem.

## Summary table
| # | Concept | In the code |
|---|---|---|
| 1 | IP / localhost | `-addr localhost:...` |
| 2 | Port | `5001`, `5002`, `5003` |
| 3 | TCP | `"tcp"` in Listen / Dial |
| 4 | Client and server | Each Node plays both roles |
| 5 | Listen and accept | `Start()`, `acceptLoop()` |
| 6 | Dial | `call()` → `DialTimeout` |
| 7 | Connection | `conn`, `Close()` |
| 8 | RPC + serialization | `Raft.Ping`, `PingArgs`, `PingReply` |
| 9 | Timeouts / failures | `DialTimeout`, `time.After` |

## How a single ping travels
```
Server 1 (client)                          Server 2 (server)
─────────────────                          ─────────────────
                                           net.Listen(:5002)  ← waiting
DialTimeout(:5002)  ──── TCP connect ────► Accept()
                                           ServeConn(conn)
client.Go("Raft.Ping", {From:1})
      │  serialize → bytes
      └──────────── request ─────────────► decode → run Ping()
                                           reply.From = 2
      ┌──────────── reply ◄─────────────── serialize → bytes
decode → reply.From = 2
logf("received ping reply from server 2")
client.Close()      ──── TCP close ──────►
```
