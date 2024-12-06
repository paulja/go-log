# go-log

A distributed log implementation written in Go. The service uses gRPC for all external APIs, Serf and Raft with a multiplex implementation for inter-server communication, replication and resilience. Fully observable via structured logging and traces.

---

Developed while reading the book Distributed Services in Go by Travis Jeffery.
