## EchoAnyPort

Use bpf socket steering to route tcp connections to any local addr/ports to a single listening socket. 
- Original intro: https://github.com/jsitnicki/ebpf-summit-2020. 
- Stable toolset: https://github.com/cloudflare/tubular . 

With tubular, you can bind/rebind a service on the fly to another port, subset of ports, all ports, and subnets of addresses, without having to restart the running service.

Here we try use tubular in code to map a range of pre-specified addr/ports to a single listening socket.

### Install requirements:

- have linux version >=v5.10, and bpf (bpftool, libbpf) installed.
- install tubular: go install github.com/cloudflare/tubular/cmd/tubectl@latest

### Build

go build ./cmd/tube_echo

### Run

go run ./cmd/tube_echo 5001 5002 ... (all ports whose conns go to a single listening socket)

### Test

connect to any local addresses with specified ports:

- nc localhost 5001
- nc 127.0.0.12 5002
- nc 127.0.0.101 5001

