## EchoAnyPort

Use bpf socket steering to route tcp connections to any ports to a single listening socket. 
- Original intro: https://github.com/jsitnicki/ebpf-summit-2020. 
- Mature wrapper: https://github.com/cloudflare/tubular . 

Here we can use tubular wrapper.

### Install requirements:

- have linux version >=v5.10, and bpf (bpftool, libbpf) installed.
- install tubular: go install github.com/cloudflare/tubular/cmd/tubectl@latest

### Build

go build ./cmd/tube_echo

### Run

go run ./cmd/tube_echo 5001 5002 ... (all ports whose conns go to a single listener)

### Test

nc localhost 5001
