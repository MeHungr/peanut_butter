![Go Version](https://img.shields.io/badge/Go-1.23-blue)
![Build](https://github.com/MeHungr/peanut_butter/actions/workflows/ci.yml/badge.svg)

# Peanut Butter

**Peanut Butter** is a simple HTTPS command-and-control (C2) framework built for red vs. blue team competitions and personal learning.
It includes a server, agent, and CLI (`pbctl`) with SQLite-backed data persistence. Peanut Butter agents communicate with the server on a beaconing system by sending POST requests to various endpoints on a preconfigured callback interval.
The C2 has an agent and server written for linux, windows, freebsd, and macos (untested).

## Features
- SQLite storage with `agents`, `tasks`, and `results` tables
- Server API with endpoints for agent registration, tasks, results, and targets
- Cross-platform agent that registers, polls tasks, and sends results
- CLI (`pbctl`) for managing agents, targets, and tasks
- Status tracking of agents (online, stale, offline) with humanized `last seen` value
- Pretty CLI table output with customizable output specifications
---

## Project Structure
```
cmd
├── server # main entrypoint for the C2 server
├── agent # agent entrypoint
└── cli # cobra-based CLI
internal
├── agent # agent runtime
├── api # shared API types
├── cli # cli & server communication logic
├── conversion # conversion between api and storage types
├── pberrors # errors used internally
├── server # HTTP handlers & server logic
├── storage # database access (SQLite + sqlx)
├── transport # Defines interfaces and implementations for client-server communcation
├── ui # pretty-printing helpers
└── util # Utilities and helper functions
```

## Installation
Clone the repo, and configure before building
1. Clone the repo and cd into the root directory
``` bash
git clone https://github.com/MeHungr/peanut_butter
cd peanut_butter
```
2. Edit `cmd/agent/main.go`
- Set agentID, serverIP, serverPort, callbackInterval, and debugMode
3. Edit `cmd/server/main.go`
- Set port
4. Edit `cmd/cli/main.go`
- Set baseURL
5. Generate server TLS certificates (AGENTS WILL NEED TO BE REBUILT AND REDEPLOYED AFTER RUNNING THIS)
``` bash
make build-certs
```
6. Build the binaries
``` bash
make # places binaries in <project root>/bin and installs the cli
```

## Usage
### Starting the server and agents
Both the server and agent can now be installed as services on linux, windows, macos, and freebsd.
Simply add a subcommand when running the executable:
``` bash
pbagent install   | pbserver install
pbagent uninstall | pbserver uninstall
pbagent start     | pbserver start
pbagent stop      | pbserver stop
```
To run the executables in the foreground, run them with no subcommands:
``` bash
pbagent
pbserver
```
### Interact using the CLI
``` bash
pbctl help
pbctl agents list
pbctl targets set 10.1.1.1 10.1.1.2
pbctl command "echo Hello"
```

## Roadmap
[TODO](TODO.md)

# Disclaimer
⚠️ **For educational use and red team competitions only. Do not use this tool for unauthorized or illegal activity.**
