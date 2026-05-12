# portscan

![CI](https://github.com/MaplesMcDepth/portscan/actions/workflows/ci.yml/badge.svg)
![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)


Fast TCP port scanner — know what's listening.

## Install

```bash
go install github.com/MaplesMcDepth/portscan/cmd/portscan@latest
```

## Commands

### Scan common ports
```bash
portscan localhost
portscan 192.168.1.1
```

### Scan port range
```bash
portscan -p 1-1000 example.com
portscan -p 22,80,443,8080 localhost
```

### JSON output
```bash
portscan -j -p 1-100 localhost
```

### Fast scan with more workers
```bash
portscan -c 200 -t 500 -p 1-1000 192.168.1.1
```

## Options

| Flag | Description |
|------|-------------|
| `-p string` | Ports: `common`, `all`, or range (default `common`) |
| `-t int` | Timeout in ms (default 1000) |
| `-c int` | Concurrency (default 100) |
| `-j` | JSON output |
| `-v` | Verbose (show closed) |
| `-open` | Show only open ports |
