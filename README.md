# sshpar

Parallel script runner over SSH written in Go. Lightweight, fast, and Docker-ready.

## Features

- Runs shell scripts on multiple servers in parallel
- Simple configuration via YAML
- Script templating from `templates/` folder
- Logs all output into a single file
- Docker support for easy packaging

## Requirements

- Docker
- Go (if building manually)

## Configuration

Edit `config.yaml`:

```yaml
script: restart_nginx.sh        # script file from templates/
password: yourpassword          # SSH password
hosts_file: hosts.txt
log_file: output.log
```

## Hosts File Format

```
root@192.168.1.10
root@192.168.1.11:2222
```

## Run with Docker

```bash
docker build -t sshpar .
docker run --rm \
  -v "$PWD/templates:/app/templates" \
  -v "$PWD/config.yaml:/app/config.yaml" \
  -v "$PWD/hosts.txt:/app/hosts.txt" \
  -v "$PWD/output.log:/app/output.log" \
  sshpar
```

## Output

All results go to `output.log` with per-host headers.

## Warning

This version uses `InsecureIgnoreHostKey()` for simplicity. Use carefully or add secure host key verification in production.
