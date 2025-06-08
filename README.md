# lhcli - Longhorn CLI

A comprehensive command-line interface for managing Longhorn storage system.

## Features

- Complete volume lifecycle management
- Snapshot and backup operations
- Node and disk management
- Real-time monitoring
- Multiple output formats (table, JSON, YAML)
- Context-based configuration
- Batch operations support

## Installation

```bash
go install github.com/longhorn/lhcli@latest
```

Or build from source:

```bash
git clone https://github.com/longhorn/lhcli.git
cd lhcli
go build -o lhcli main.go
```

## Configuration

Create a configuration file at `~/.lhcli/config.yaml`:

```yaml
contexts:
  - name: production
    endpoint: https://longhorn.prod.example.com
    namespace: longhorn-system
    auth:
      type: kubeconfig
      path: ~/.kube/config

current-context: production

defaults:
  output-format: table
  confirmation: true
  timeout: 30s
```

## Usage

### Volume Management

```bash
# List all volumes
lhcli volume list

# Create a volume
lhcli volume create my-volume --size 10Gi --replicas 3

# Get volume details
lhcli volume get my-volume

# Delete a volume
lhcli volume delete my-volume
```

### Snapshot Management

```bash
# Create a snapshot
lhcli snapshot create my-volume --name my-snapshot

# List snapshots
lhcli snapshot list my-volume

# Revert to snapshot
lhcli snapshot revert my-volume my-snapshot
```

### Backup Management

```bash
# Create a backup
lhcli backup create my-volume

# List backups
lhcli backup list

# Restore from backup
lhcli backup restore <backup-url> --volume new-volume
```

### Monitoring

```bash
# Monitor volumes in real-time
lhcli monitor volumes --interval 5s

# Monitor nodes
lhcli monitor nodes

# Follow events
lhcli monitor events --follow
```

### Troubleshooting

```bash
# Run troubleshooting checks
lhcli troubleshoot
```

This command inspects nodes, replicas and disks to report common issues such as orphaned replicas or low disk space.

## Development

### Building

```bash
go build -o lhcli main.go
```

### Testing

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Apache License 2.0
