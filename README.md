![Go Version](https://img.shields.io/badge/Go-1.23-blue)
![Build](https://github.com/MeHungr/peanut_butter/actions/workflows/ci.yml/badge.svg)

# Peanut Butter

**Peanut Butter** is a simple HTTP command-and-control (C2) framework built for red vs. blue team competitions and personal learning.
It includes a server, agent, and CLI (`pbctl`) with SQLite-backed data persistence. Peanut Butter agents communicate with the server on a beaconing system by sending POST requests to various endpoints on a preconfigured callback interval.

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
├── api # shared API types
├── server # HTTP handlers & server logic
├── storage # database access (SQLite + sqlx)
├── agent # agent runtime
├── cli # cli & server communication logic
└── ui # pretty-printing helpers
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
4. Build the binaries
``` bash
make # places binaries in <project root>/bin and installs the cli
```

## Usage
### Starting the server and agents
Currently, the server and agents run as executable files. Execute the respective file to start the program.
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
