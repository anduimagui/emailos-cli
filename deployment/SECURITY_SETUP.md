# 🔒 EmailOS Server Security Configuration

## Security Measures Implemented

### 1. SSH Hardening ✅
- **Key Type:** ED25519 (most secure)
- **Root Login:** Disabled
- **Password Authentication:** Disabled
- **Only User Allowed:** emailos-admin
- **Max Auth Tries:** 3
- **Login Grace Time:** 30 seconds
- **Strong Ciphers Only:** ChaCha20-Poly1305, AES256-GCM
- **Protocol:** SSH v2 only

### 2. Firewall (UFW) ✅
- **Default Policy:** Deny all incoming, allow outgoing
- **Open Ports:**
  - 22/tcp (SSH)
  - 60000-61000/udp (Mosh)
  - 8080/tcp (Web Server)
  - 8089/tcp (EmailOS API)

### 3. Intrusion Prevention (Fail2ban) ✅
- **SSH Protection:** 3 failed attempts = 1 hour ban
- **Aggressive Mode:** 2 failed attempts in 1 hour = 24 hour ban
- **Monitors:** /var/log/auth.log

### 4. User Security ✅
- **Admin User:** emailos-admin (sudo privileges)
- **Root Access:** Disabled for direct login
- **Sudo:** Passwordless for emailos-admin

### 5. Automatic Updates ✅
- **Security Updates:** Automatic installation
- **Kernel Updates:** Automatic download
- **Reboot:** Manual (notification only)

### 6. Port Knocking ✅
- **Sequence:** 7000, 8000, 9000
- **Timeout:** 10 seconds
- **Purpose:** Additional SSH protection layer

### 7. System Hardening ✅
- **IPv6:** Disabled
- **IP Spoofing Protection:** Enabled
- **ICMP Redirects:** Ignored
- **Source Routing:** Disabled
- **SYN Flood Protection:** Enabled
- **Martian Packets:** Logged

## 🔐 Access Instructions

### Primary Access (SSH)
```bash
# Connect as emailos-admin (not root!)
ssh -i ~/.ssh/emailos_server_key emailos-admin@46.62.173.41

# If using port knocking (optional extra security)
knock 46.62.173.41 7000 8000 9000
ssh -i ~/.ssh/emailos_server_key emailos-admin@46.62.173.41
```

### Mobile Access (Mosh)
```bash
mosh --ssh="ssh -i ~/.ssh/emailos_server_key" emailos-admin@46.62.173.41
```

### Termius Configuration
- **Host:** 46.62.173.41
- **Username:** emailos-admin (NOT root)
- **Port:** 22
- **Key:** Contents of `~/.ssh/emailos_server_key`

## 🚨 Important Security Notes

1. **NEVER share your private key** (`~/.ssh/emailos_server_key`)
2. **Root login is disabled** - Always use emailos-admin
3. **Use sudo for admin tasks** - `sudo command`
4. **Monitor logs regularly** - Check `/var/log/auth.log`
5. **Keep key secure** - Set permissions to 600

## 📊 Security Status Commands

Run these as emailos-admin:

```bash
# Check fail2ban status
sudo fail2ban-client status sshd

# View blocked IPs
sudo iptables -L -n

# Check firewall status
sudo ufw status verbose

# View SSH attempts
sudo tail -f /var/log/auth.log

# Check listening ports
sudo ss -tlnp

# System security audit
sudo lynis audit system  # (install with: sudo apt install lynis)
```

## 🔄 Maintenance Tasks

### Weekly
- Review auth logs: `sudo grep "Failed password" /var/log/auth.log`
- Check fail2ban: `sudo fail2ban-client status`

### Monthly
- Update system: `sudo apt update && sudo apt upgrade`
- Review open ports: `sudo ss -tlnp`
- Check disk usage: `df -h`

### As Needed
- Unban IP: `sudo fail2ban-client unban <IP>`
- Add firewall rule: `sudo ufw allow from <IP> to any port 22`
- Restart SSH: `sudo systemctl restart ssh`

## 🚫 What's Blocked

- All incoming traffic except specified ports
- Direct root login
- Password authentication
- Weak SSH ciphers
- IPv6 traffic
- ICMP redirects
- Source routing
- Multiple failed login attempts

## ✅ Security Checklist

- [x] SSH key-only authentication
- [x] Non-root sudo user
- [x] Firewall enabled
- [x] Fail2ban active
- [x] Automatic security updates
- [x] Port knocking available
- [x] System hardening applied
- [x] Strong encryption only
- [x] Login monitoring enabled

## 🆘 Emergency Access

If locked out:

1. **Use Hetzner Console** (web-based console in Hetzner Cloud panel)
2. **Reset via Hetzner API:**
   ```bash
   hcloud server reset emailos-server
   ```
3. **Boot into rescue mode:**
   ```bash
   hcloud server enable-rescue emailos-server
   hcloud server reboot emailos-server
   ```

## 📝 Security Log

- **2025-08-13:** Initial security setup completed
  - SSH hardened with ED25519 keys
  - Root login disabled
  - Firewall configured
  - Fail2ban installed
  - Automatic updates enabled
  - Port knocking configured
  - System hardening applied

---

**Server is now secured according to best practices!**
Only you can access it with your ED25519 key as emailos-admin user.