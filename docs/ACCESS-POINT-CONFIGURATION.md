# Access Point Configuration Guide

This guide provides step-by-step instructions for configuring **Ruckus**, **Aruba**, and **UniFi** access points to work with the FreeRADIUS Google SSO authentication system.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [RADIUS Server Information](#radius-server-information)
- [VLAN Configuration Overview](#vlan-configuration-overview)
- [Ruckus Configuration](#ruckus-configuration)
- [Aruba Configuration](#aruba-configuration)
- [UniFi Configuration](#unifi-configuration)
- [Testing and Troubleshooting](#testing-and-troubleshooting)

---

## Prerequisites

Before configuring your access points, ensure you have:

1. **FreeRADIUS server** running and accessible from your network
2. **RADIUS shared secret** (default: `testing123`, change in production)
3. **VLAN IDs** configured on your network switches
4. **Network connectivity** between APs and RADIUS server
5. **Admin access** to your AP controller/management interface

---

## RADIUS Server Information

### Server Details

```
RADIUS Server IP: <YOUR_RADIUS_SERVER_IP>
Authentication Port: 1812
Accounting Port: 1813
Shared Secret: testing123 (CHANGE THIS!)
Authentication Method: EAP-PEAP / MSCHAPv2
```

### VLAN Assignment

The system dynamically assigns VLANs based on user type:

| User Type | VLAN ID | Description |
|-----------|---------|-------------|
| Faculty   | 10      | Faculty members |
| Student   | 20      | Students |
| Staff     | 30      | Administrative staff |
| Guest     | 40      | Guest users |

---

## Ruckus Configuration

### Using Ruckus SmartZone / ZoneDirector

#### 1. Configure RADIUS Server

**Path:** `Configuration > Authentication > RADIUS`

1. Click **Create New** RADIUS server
2. Enter the following details:
   - **Name:** `FreeRADIUS-GoogleSSO`
   - **IP Address:** `<YOUR_RADIUS_SERVER_IP>`
   - **Port:** `1812`
   - **Shared Secret:** `testing123` (change to your secret)
   - **Accounting Port:** `1813`
   - **Enable Accounting:** ✅ Checked

3. Click **OK** to save

#### 2. Create WLAN/SSID

**Path:** `Configuration > WLANs`

1. Click **Create New**
2. Configure basic settings:
   - **WLAN Name:** `YourNetwork`
   - **SSID:** `YourNetwork`
   - **Authentication Method:** `WPA2 Enterprise`
   - **Encryption:** `AES`

3. Configure authentication:
   - **Authentication Server:** Select `FreeRADIUS-GoogleSSO`
   - **Accounting Server:** Select `FreeRADIUS-GoogleSSO`

4. Configure VLAN settings:
   - **Access VLAN:** `Use RADIUS Attributes`
   - **Enable Dynamic VLAN:** ✅ Checked

5. Advanced settings:
   - **Enable 802.11r (Fast Roaming):** ✅ Recommended
   - **Enable 802.11k/v:** ✅ Recommended for better roaming

6. Click **OK** to save

#### 3. Configure NAS Identifier (Optional)

**Path:** `Configuration > Access Points > [Select AP] > Radio`

1. Set **NAS-Identifier:** `Ruckus-AP-Building1` (customize per location)

#### 4. Apply Configuration

1. Navigate to `Configuration > WLANs`
2. Select your WLAN
3. Click **Apply** to push configuration to APs

---

## Aruba Configuration

### Using Aruba AirWave / Instant / Mobility Controller

#### 1. Configure RADIUS Server (Aruba Instant)

**Path:** `System > External Services > RADIUS Authentication`

1. Click **+** to add new server
2. Enter server details:
   ```
   Server IP Address: <YOUR_RADIUS_SERVER_IP>
   Auth Port: 1812
   Acct Port: 1813
   Shared Secret: testing123
   ```

3. Enable options:
   - **RFC 3576 (Dynamic Authorization):** Enabled
   - **Accounting:** Enabled
   - **Interim Accounting Interval:** 300 seconds

4. Click **OK**

#### 2. Configure RADIUS Server (Aruba Controller)

**CLI Commands:**

```bash
# Configure RADIUS server
aaa authentication-server radius FreeRADIUS-GoogleSSO
    host <YOUR_RADIUS_SERVER_IP>
    key testing123
    acct-port 1813
    auth-port 1812
!

# Configure server group
aaa server-group FreeRADIUS-Group
    auth-server FreeRADIUS-GoogleSSO
    set role condition role value-of Filter-Id
!

# Enable dynamic authorization
aaa rfc-3576-server <YOUR_RADIUS_SERVER_IP>
aaa rfc-3576-server key testing123
```

#### 3. Create Virtual AP/SSID

**Aruba Instant - Path:** `Networks > Employee Network`

1. Click **+** to create new network
2. Configure:
   - **Network Name (SSID):** `YourNetwork`
   - **Client IP Assignment:** `Virtual Controller Managed`
   - **Security Level:** `WPA2/WPA3 Enterprise`
   - **Authentication Servers:** Select `FreeRADIUS-GoogleSSO`

3. VLAN Configuration:
   - **Client VLAN Assignment:** `Enabled`
   - **VLAN:** `Use RADIUS Derived`
   - **Default VLAN ID:** `99` (fallback VLAN)

4. Enable accounting:
   - **Accounting:** ✅ Enabled
   - **Accounting Server:** Select `FreeRADIUS-GoogleSSO`

**Aruba Controller - CLI Commands:**

```bash
# Create SSID profile
wlan ssid-profile YourNetwork
    essid YourNetwork
    opmode wpa2-aes
!

# Create VLAN profile
wlan virtual-ap YourNetwork
    aaa-profile FreeRADIUS-AAA
    ssid-profile YourNetwork
    vlan-mode derived
    default-vlan 99
!

# AAA profile
aaa profile FreeRADIUS-AAA
    authentication-dot1x FreeRADIUS-Group
    dot1x-default-role authenticated
    dot1x-server-group FreeRADIUS-Group
    accounting FreeRADIUS-Group
!
```

#### 4. Configure Role Mapping

**Path:** `Configuration > User Roles`

Create roles for each VLAN:

```bash
user-role faculty
    access-list session allowall
!

user-role student
    access-list session allowall
!

user-role staff
    access-list session allowall
!

user-role guest
    access-list session guest-acl
!
```

---

## UniFi Configuration

### Using UniFi Network Controller

#### 1. Configure RADIUS Profile

**Path:** `Settings > Profiles > RADIUS`

1. Click **Create New RADIUS Profile**
2. Configure:
   - **Profile Name:** `FreeRADIUS-GoogleSSO`
   - **Authentication Servers:**
     - **IP Address:** `<YOUR_RADIUS_SERVER_IP>`
     - **Port:** `1812`
     - **Password:** `testing123`
   - **Accounting Servers:**
     - **IP Address:** `<YOUR_RADIUS_SERVER_IP>`
     - **Port:** `1813`
     - **Password:** `testing123`

3. Advanced settings (optional):
   - **Interim Update Interval:** `300` seconds
   - **Enable RADIUS Assigned VLANs:** ✅ Enabled

4. Click **Save**

#### 2. Create WiFi Network (SSID)

**Path:** `Settings > WiFi > Create New WiFi Network`

1. Configure basic settings:
   - **Name/SSID:** `YourNetwork`
   - **Enabled:** ✅ Checked
   - **Security Protocol:** `WPA2 Enterprise` or `WPA3 Enterprise`

2. Configure RADIUS:
   - **RADIUS Profile:** Select `FreeRADIUS-GoogleSSO`
   - **RADIUS MAC Authentication:** ❌ Disabled (unless needed)

3. Configure VLAN:
   - **Network:** `Use RADIUS Assigned VLAN`
   - **Fallback VLAN:** `99` (optional, for users without VLAN assignment)

4. Advanced settings:
   - **Fast Roaming (802.11r):** ✅ Enabled (recommended)
   - **BSS Transition (802.11v):** ✅ Enabled
   - **Proxy ARP:** ✅ Enabled (optional)
   - **Group Rekey Interval:** `3600` seconds

5. Click **Save**

#### 3. Configure VLANs on Network

**Path:** `Settings > Networks`

Ensure VLANs are configured:

1. Create networks for each VLAN:
   - **Faculty Network:** VLAN ID `10`
   - **Student Network:** VLAN ID `20`
   - **Staff Network:** VLAN ID `30`
   - **Guest Network:** VLAN ID `40`

2. For each network, configure:
   - **DHCP Mode:** `DHCP Server` or `DHCP Relay`
   - **Gateway/Subnet:** Configure as per your network design
   - **DHCP Range:** Appropriate IP range for each VLAN

#### 4. Configure Switch Port Profiles (if using UniFi switches)

**Path:** `Settings > Profiles > Switch Port`

1. Create a trunk port profile for APs:
   - **Name:** `AP-Trunk-Port`
   - **Native VLAN:** `Management VLAN` (e.g., VLAN 1)
   - **Tagged VLANs:** `10, 20, 30, 40` (all user VLANs)
   - **Port Isolation:** ❌ Disabled

2. Apply this profile to ports where APs are connected

---

## Advanced Configuration

### 1. Enable RADIUS Attributes

Ensure your RADIUS server sends these attributes:

```
Tunnel-Type = VLAN (13)
Tunnel-Medium-Type = IEEE-802 (6)
Tunnel-Private-Group-ID = <VLAN_ID>
Filter-Id = <USER_TYPE>
```

### 2. NAS-Identifier Configuration

Configure unique NAS-Identifier for each AP or location to track authentication requests:

**Ruckus:**
```
Per AP: Configuration > Access Points > [AP] > Radio > NAS-ID
```

**Aruba:**
```bash
ap-name <AP-NAME>
    nas-id "Building1-Floor2-AP01"
!
```

**UniFi:**
- UniFi automatically uses AP MAC address as NAS-Identifier

### 3. Accounting Configuration

Enable accounting to track user sessions:

**Accounting Interval:** 300 seconds (recommended)

**Accounting Attributes to Send:**
- Acct-Session-Id
- Acct-Session-Time
- Acct-Input-Octets
- Acct-Output-Octets
- Acct-Terminate-Cause

### 4. Certificate Configuration (for secure PEAP)

For production environments, install proper SSL certificates on RADIUS server:

1. Generate/obtain SSL certificate for RADIUS server
2. Configure FreeRADIUS with certificate:
   ```
   File: /etc/freeradius/3.0/certs/server.pem
   ```
3. Update certificate common name (CN) to match RADIUS server hostname
4. Distribute CA certificate to client devices (optional but recommended)

---

## Testing and Troubleshooting

### 1. Test RADIUS Connectivity

From AP or controller, test RADIUS server connectivity:

**Using radtest (Linux):**
```bash
radtest testuser testpass <RADIUS_IP>:1812 0 testing123
```

**Expected Response:**
```
Received Access-Accept packet from <RADIUS_IP>:1812
```

### 2. Monitor RADIUS Logs

On RADIUS server:
```bash
# View authentication logs
docker exec radius-server tail -f /var/log/freeradius/radius.log

# View post-authentication logs
docker exec radius-mysql mysql -u radius -p -e "SELECT * FROM radius.radpostauth ORDER BY id DESC LIMIT 10;"
```

### 3. Common Issues

#### Issue: Authentication fails with "Access-Reject"

**Check:**
- User exists in Google Workspace with correct domain
- RADIUS shared secret matches on AP and server
- RADIUS server is reachable (ping test)
- Check `/var/log/freeradius/radius.log` for errors

**Debug:**
```bash
docker exec radius-server radtest <username>@krea.edu.in <password> localhost 0 testing123
```

#### Issue: User authenticates but no VLAN assignment

**Check:**
- RADIUS server returns `Tunnel-Private-Group-ID` attribute
- VLAN ID exists on network switches
- AP/Controller has "RADIUS VLAN assignment" enabled
- Check radpostauth table for VLAN assignment:
  ```sql
  SELECT username, vlan_id, reply FROM radpostauth WHERE username = 'user@domain.com' ORDER BY authdate DESC LIMIT 1;
  ```

#### Issue: Intermittent connection drops

**Check:**
- Accounting interim interval not too frequent (use 300s)
- RADIUS server not overloaded (check CPU/memory)
- Network latency between AP and RADIUS server
- Enable fast roaming (802.11r) on SSID

### 4. Debug Mode

Enable debug mode on RADIUS server:

```bash
# Stop RADIUS server
docker exec radius-server sv stop radiusd

# Start in debug mode
docker exec -it radius-server radiusd -X

# Test authentication and observe detailed logs
```

### 5. Verify VLAN Assignment

**Check client VLAN assignment:**

**Ruckus:**
```
Path: Monitor > Clients > [Select Client] > Details
Look for: VLAN ID
```

**Aruba:**
```bash
show user-table
```

**UniFi:**
```
Path: Clients > [Select Client] > Details
Look for: Network (shows VLAN assignment)
```

### 6. Packet Capture

Capture RADIUS traffic for detailed analysis:

```bash
# On RADIUS server
docker exec radius-server tcpdump -i any -n port 1812 or port 1813 -w /tmp/radius.pcap

# Download and analyze with Wireshark
docker cp radius-server:/tmp/radius.pcap ./radius.pcap
```

---

## Security Best Practices

### 1. Change Default Shared Secret

Update RADIUS shared secret in:
- `.env` file: `RADIUS_SECRET=<strong-random-secret>`
- All AP configurations

Generate strong secret:
```bash
openssl rand -base64 32
```

### 2. Restrict RADIUS Server Access

Configure firewall to allow only AP IPs:

```bash
# Example iptables rules
iptables -A INPUT -p udp --dport 1812 -s <AP_SUBNET> -j ACCEPT
iptables -A INPUT -p udp --dport 1813 -s <AP_SUBNET> -j ACCEPT
iptables -A INPUT -p udp --dport 1812 -j DROP
iptables -A INPUT -p udp --dport 1813 -j DROP
```

### 3. Configure Client Certificate Validation (Optional)

For high-security environments, require client certificates in addition to username/password.

### 4. Enable RADIUS Message-Authenticator

Ensure Message-Authenticator attribute is present in all RADIUS packets to prevent spoofing.

### 5. Monitor Authentication Logs

Regularly review authentication logs for:
- Failed login attempts
- Unusual VLAN assignments
- Authentication from unexpected NAS devices

Access admin dashboard: `https://admin-radius.krea.edu.in/`

---

## Quick Reference Table

| Vendor | RADIUS Config Path | VLAN Assignment Setting | Notes |
|--------|-------------------|------------------------|-------|
| **Ruckus** | Configuration > Authentication > RADIUS | Access VLAN > "Use RADIUS Attributes" | Enable "Dynamic VLAN" |
| **Aruba** | System > External Services > RADIUS | Client VLAN Assignment > "Use RADIUS Derived" | Set default/fallback VLAN |
| **UniFi** | Settings > Profiles > RADIUS | Network > "Use RADIUS Assigned VLAN" | Configure VLANs in Networks first |

---

## Support and Resources

### FreeRADIUS Documentation
- [FreeRADIUS Wiki](https://wiki.freeradius.org/)
- [RADIUS Configuration](https://wiki.freeradius.org/config/Configuration-files)

### Vendor Documentation
- [Ruckus RADIUS Configuration](https://docs.ruckuswireless.com/)
- [Aruba 802.1X Configuration](https://www.arubanetworks.com/techdocs/)
- [UniFi RADIUS Setup](https://help.ui.com/hc/en-us/articles/360006615434)

### Admin Dashboard
- URL: `https://admin-radius.krea.edu.in/`
- View authentication logs, user sessions, and reports

### Captive Portal
- URL: `https://radius.krea.edu.in/`
- User authentication and session management

---

## Appendix: Example RADIUS Clients Configuration

Add AP/Controller IP addresses to RADIUS server:

**File:** `freeradius/config/clients.conf`

```
# Ruckus Zone Director
client ruckus-zd {
    ipaddr = 192.168.1.10
    secret = testing123
    shortname = ruckus-zd
    nas_type = other
}

# Aruba Mobility Controller
client aruba-mc {
    ipaddr = 192.168.1.20
    secret = testing123
    shortname = aruba-mc
    nas_type = other
}

# UniFi Controller
client unifi-controller {
    ipaddr = 192.168.1.30
    secret = testing123
    shortname = unifi-controller
    nas_type = other
}

# Allow entire AP subnet (not recommended for production)
client ap-subnet {
    ipaddr = 192.168.100.0/24
    secret = testing123
    shortname = ap-network
    nas_type = other
}
```

After updating, restart RADIUS server:
```bash
docker-compose restart freeradius
```

---

**Document Version:** 1.0
**Last Updated:** January 2026
**Maintained by:** Network Administration Team
