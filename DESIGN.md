# Longhorn CLI (lhcli) Design Document

## Overview

`lhcli` is a command-line interface tool for Longhorn, providing comprehensive management capabilities that match and extend the Longhorn WebUI functionality.

## Architecture

### Core Components

```
lhcli/
                              pkg/
client/            # Longhorn API client wrapper
formatter/         # Output formatting (table, json, yaml)
utils/             # Common utilities
config/            # Configuration management
monitor/           # Monitoring utilities
      main.go
```

## Command Structure

### 1. Volume Management

```bash
# List volumes
lhcli volume list [--namespace <ns>] [--format json|yaml|table]

# Create volume
lhcli volume create <name> --size <size> [--replicas <count>] [--frontend <type>]

# Get volume details
lhcli volume get <name> [--detailed]

# Delete volume
lhcli volume delete <name> [--force]

# Attach/Detach volume
lhcli volume attach <name> --node <node-name>
lhcli volume detach <name>

# Update volume
lhcli volume update <name> [--replicas <count>] [--access-mode <mode>]

# Volume operations
lhcli volume expand <name> --size <new-size>
lhcli volume activate <name>
lhcli volume deactivate <name>
```

### 2. Snapshot Management

```bash
# Create snapshot
lhcli snapshot create <volume-name> --name <snapshot-name> [--labels key=value]

# List snapshots
lhcli snapshot list <volume-name>

# Delete snapshot
lhcli snapshot delete <volume-name> <snapshot-name>

# Revert to snapshot
lhcli snapshot revert <volume-name> <snapshot-name>

# Clone from snapshot
lhcli snapshot clone <volume-name> <snapshot-name> --new-volume <name>
```

### 3. Backup Management

```bash
# Configure backup target
lhcli backup target set <backup-target-url> [--credential-secret <name>]
lhcli backup target get

# Create backup
lhcli backup create <volume-name> [--snapshot <name>] [--labels key=value]

# List backups
lhcli backup list [--volume <name>]

# Restore from backup
lhcli backup restore <backup-url> --volume <name>

# Delete backup
lhcli backup delete <backup-url>
```

### 4. Node Management

```bash
# List nodes
lhcli node list [--format json|yaml|table]

# Get node details
lhcli node get <node-name>

# Enable/Disable scheduling
lhcli node scheduling enable <node-name>
lhcli node scheduling disable <node-name>

# Evict replicas
lhcli node evict <node-name> [--force]

# Update node tags
lhcli node tag add <node-name> <tag>
lhcli node tag remove <node-name> <tag>

# Disk management
lhcli node disk add <node-name> --path <path> --storage <size>
lhcli node disk remove <node-name> <disk-id>
lhcli node disk update <node-name> <disk-id> [--tags <tags>]
```

### 5. Engine & Replica Management

```bash
# List engines
lhcli engine list [--volume <name>]

# Get engine details
lhcli engine get <engine-name>

# List replicas
lhcli replica list [--volume <name>] [--node <name>]

# Get replica details
lhcli replica get <replica-name>

# Delete replica
lhcli replica delete <replica-name>
```

### 6. System Settings

```bash
# Get all settings
lhcli settings list

# Get specific setting
lhcli settings get <setting-name>

# Update setting
lhcli settings update <setting-name> --value <value>

# Reset to default
lhcli settings reset <setting-name>
```

### 7. Monitoring & Diagnostics

```bash
# Real-time monitoring
lhcli monitor volumes [--interval 5s]
lhcli monitor nodes [--interval 5s]
lhcli monitor events [--follow]

# Generate support bundle
lhcli support bundle generate [--output <path>]

# Health check
lhcli health check [--detailed]

# Performance metrics
lhcli metrics volume <name>
lhcli metrics node <name>
```

### 8. Advanced Features

```bash
# Recurring jobs
lhcli recurring-job create <name> --cron <expression> --task <snapshot|backup>
lhcli recurring-job list
lhcli recurring-job delete <name>

# DR volumes
lhcli dr volume create <name> --backup-target <url>
lhcli dr volume activate <name>

# System backup/restore
lhcli system backup create
lhcli system restore <backup-name>

# Batch operations
lhcli batch volume create --file <volumes.yaml>
lhcli batch volume delete --selector <label-selector>
```

## Configuration

### Configuration File

```yaml
# ~/.lhcli/config.yaml
contexts:
 - name: production
   endpoint: https://longhorn.prod.example.com
   namespace: longhorn-system
   auth:
     type: kubeconfig
     path: ~/.kube/config
 - name: staging
   endpoint: https://longhorn.staging.example.com
   namespace: longhorn-system
   auth:
     type: token
     token: <bearer-token>

current-context: production

defaults:
 output-format: table
 confirmation: true
 timeout: 30s
```

### Global Flags

```bash
--context <name>           # Use specific context
--namespace <ns>           # Override namespace
--output, -o <format>      # Output format (table|json|yaml|wide)
--no-headers              # Don't print headers (table format)
--dry-run                 # Preview actions without executing
--verbose, -v             # Verbose output
--quiet, -q               # Minimal output
--config <path>           # Config file path
--timeout <duration>      # Request timeout
```

## Implementation Details

### API Client

```go
// pkg/client/client.go
type Client struct {
   config     *Config
   httpClient *http.Client
   baseURL    string
}

func (c *Client) Volumes() VolumeInterface
func (c *Client) Nodes() NodeInterface
func (c *Client) Backups() BackupInterface
// ... other resources
```

### Output Formatting

```go
// pkg/formatter/formatter.go
type Formatter interface {
   Format(data interface{}) ([]byte, error)
}

type TableFormatter struct {
   columns []Column
   options TableOptions
}

type JSONFormatter struct {
   pretty bool
}

type YAMLFormatter struct{}
```

### Interactive Features

- Confirmation prompts for destructive operations
- Progress bars for long-running operations
- Auto-completion for resource names
- Interactive selection for batch operations

### Error Handling

- Clear, actionable error messages
- Retry logic for transient failures
- Graceful degradation when API features are unavailable
- Detailed error context with --verbose flag

## Extended CLI Features (Beyond WebUI)

1. **Batch Operations**: Process multiple resources from YAML/JSON files
2. **Watch Mode**: Monitor resources for changes in real-time
3. **Scriptable Output**: Machine-readable formats for automation
4. **Shell Completion**: Bash/Zsh/Fish completion support
5. **Offline Mode**: View cached data when API is unavailable
6. **Performance Profiling**: Built-in performance analysis tools
7. **Template Support**: Generate resource definitions from templates
8. **Diff Command**: Compare configurations between environments
9. **Audit Trail**: Log all CLI operations for compliance
10. **Plugin System**: Extend functionality with custom plugins

## Testing Strategy

- Unit tests for all command logic
- Integration tests against mock Longhorn API
- E2E tests against real Longhorn cluster
- Performance benchmarks
- CLI behavior tests (flags, output formats, etc.)

## Documentation

- Comprehensive man pages
- Built-in help system with examples
- Online documentation with tutorials
- Video demonstrations
- Troubleshooting guide

## Release & Distribution

- Binary releases for Linux/macOS/Windows
- Container image with lhcli
- Homebrew formula
- APT/YUM repositories
- Kubernetes kubectl plugin compatibility
