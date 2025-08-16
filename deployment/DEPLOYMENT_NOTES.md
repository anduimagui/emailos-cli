# EmailOS Server Deployment Notes

## 🚀 Deployment Summary

Successfully deployed EmailOS to Hetzner Cloud with the following configuration:

### Server Details
- **Server Name:** emailos-server
- **IP Address:** 46.62.173.41
- **Server Type:** CPX11 (1 vCPU, 2GB RAM, 40GB SSD)
- **Location:** Helsinki (hel1)
- **OS:** Ubuntu 24.04 LTS
- **Monthly Cost:** ~€4.15/month

### Access Information

#### SSH Access
```bash
ssh -i ~/.ssh/emailos_server_key root@46.62.173.41
```

#### Mosh Access (Better for mobile/unstable connections)
```bash
mosh --ssh="ssh -i ~/.ssh/emailos_server_key" root@46.62.173.41
```

#### Web Interfaces
- General Web Server: http://46.62.173.41:8080
- EmailOS API (when configured): http://46.62.173.41:8089

### Termius Mobile Configuration
For accessing from your phone via Termius:
1. **Host:** 46.62.173.41
2. **Username:** root
3. **Port:** 22
4. **SSH Key:** Copy contents from `~/.ssh/emailos_server_key`

## 📦 Installed Software

### Core Applications
- **EmailOS (mailos):** Command-line email client installed at `/usr/local/bin/mailos`
- **Go 1.23:** Programming language runtime
- **Node.js 18:** JavaScript runtime
- **Python 3.12:** Python runtime with pipx

### AI CLI Tools
- **Claude Code:** Anthropic's AI coding assistant
  - Command: `claude` or `cc`
  - Location: `/usr/local/bin/claude`
  
- **Aider:** AI pair programming tool
  - Command: `aider` or `ai`
  - Location: `/root/.local/bin/aider`

### System Tools
- **tmux:** Terminal multiplexer for persistent sessions
- **mosh:** Mobile shell for better connectivity
- **vim:** Text editor with custom configuration
- **htop:** Interactive process viewer
- **git:** Version control system

## 🔧 Configuration Files

### Directory Structure
```
/root/
├── emailos/           # EmailOS source code
├── emailos-data/      # EmailOS data directory
├── emailos-config/    # EmailOS configuration
├── emailos-drafts/    # Email drafts storage
├── projects/          # General projects directory
├── backups/           # Backup storage
└── scripts/           # Utility scripts
    ├── setup-emailos.sh
    └── install-ai-tools.sh
```

### Custom Configurations
- **tmux:** Enhanced configuration at `/root/.tmux.conf`
- **vim:** Custom settings at `/root/.vimrc`
- **bash:** Aliases and welcome message in `/root/.bashrc`

## 📋 Quick Commands

### Server Management (from local machine)
```bash
cd /Users/andrewmaguire/LOCAL/Github/_code-main/emailos/deployment

# Server control
task status          # Show server status
task connect         # SSH to server
task mosh           # Mosh to server
task stop           # Stop server
task start          # Start server
task restart        # Restart server

# Backup and restore
task backup         # Backup server data
task restore BACKUP_FILE=./backups/emailos-backup-xxx.tar.gz

# Configuration
task termius-config  # Show Termius setup
task configure-emailos  # Configure EmailOS
```

### On the Server
```bash
# EmailOS commands
mailos setup         # Configure email account
mailos interactive   # Interactive email UI
mailos read         # Read recent emails
mailos send         # Send an email
mailos help         # Show all commands

# AI assistance
claude              # Start Claude Code
aider              # Start Aider
tmux               # Start persistent session

# Shortcuts (aliases)
m                  # mailos
mi                 # mailos interactive
cc                 # claude-code
ai                 # aider
```

## 🔄 Next Steps

1. **Configure EmailOS:**
   ```bash
   task connect
   mailos setup
   ```

2. **Set up your email account:**
   - Enter your email provider details
   - Configure app-specific password
   - Set up AI integration (optional)

3. **Start using from mobile:**
   - Import SSH key to Termius
   - Connect to server
   - Use tmux for persistent sessions
   - Run mailos or AI tools

## 🛠️ Maintenance

### Regular Tasks
- **Backup data:** Run `task backup` regularly
- **Update system:** Connect and run `apt update && apt upgrade`
- **Monitor usage:** Check server metrics in Hetzner Cloud Console

### Cost Management
- Server costs ~€4.15/month when running
- Use `task stop` when not in use to save costs
- Use `task start` to resume

## 📝 Implementation Notes

### What Was Done
1. **Created deployment infrastructure** based on vibeserver setup
2. **Deployed Ubuntu 24.04 server** to Hetzner Cloud
3. **Installed EmailOS** from local source code
4. **Configured AI tools** (Claude Code and Aider)
5. **Set up persistent environment** with tmux and mosh
6. **Created management scripts** using Task automation

### Architecture Similarities with Vibeserver
- Same Hetzner Cloud deployment approach
- Similar Taskfile.yml structure for automation
- Cloud-init for server initialization
- Mosh + tmux for mobile-friendly access
- Helsinki datacenter for low latency

### Key Differences from Vibeserver
- EmailOS-specific configuration and directories
- Additional AI tools (Aider) for code assistance
- Email-focused aliases and shortcuts
- Dedicated ports for EmailOS services

## 🚨 Troubleshooting

### Cannot connect via SSH
```bash
# Check server status
hcloud server describe emailos-server

# Verify SSH key
ls -la ~/.ssh/emailos_server_key

# Test connection with verbose output
ssh -vvv -i ~/.ssh/emailos_server_key root@46.62.173.41
```

### EmailOS not working
```bash
# Check installation
which mailos
mailos --version

# Rebuild if needed
cd /root/emailos
go build -o mailos cmd/mailos/main.go
cp mailos /usr/local/bin/
```

### AI tools not found
```bash
# Update PATH
export PATH=$PATH:/root/.local/bin
source ~/.bashrc

# Check installations
claude --version
aider --version
```

## 📞 Support

- **Hetzner Cloud Console:** https://console.hetzner.cloud/
- **EmailOS Documentation:** Available in `/root/emailos/docs/`
- **Server IP:** 46.62.173.41
- **SSH Key:** `~/.ssh/emailos_server_key`

---

*Deployment completed on: August 13, 2025*
*Server is currently: RUNNING*