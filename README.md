# Supported Platforms

1. MacOS
2. Linux

# Getting Started

## Manual run

1. `git clone https://github.com/WilsonOh/ddns-updater`
2. `make build`
3. `make init_config`
4. Update config values (cloudflare email, api token, domains and subdomains etc.) in ~/.config/ddns-updater/config.json
5. `./bin/ddns-updater`

## Setup `systemd` timer

1. Create `systemd` service unit
```bash
sudo vim /etc/systemd/system/ddns-updater.service
```
```ini
[Unit]
Description=DDNS Updater Service
After=network.target

[Service]
Type=oneshot
User=<user>
ExecStart=/usr/local/bin/ddns-updater

[Install]
WantedBy=multi-user.target
```
Set the `User` to the username of a user with permissions to run the `ddns-updater` binary
Feel free to change the description and executable file path etc.

2. Create `systemd` timer unit
```bash
sudo nvim /etc/systemd/system/ddns-updater.timer
```
```ini
[Unit]
Description=Runs DDNS Updater hourly and on boot
Requires=ddns-updater.service

[Timer]
Unit=ddns-updater.service 
OnBootSec=1min
OnCalendar=hourly
Persistent=true

[Install]
WantedBy=timers.target
```
Feel free to update the `OnCalender` field to your desired run interval

3. Start the timer
```bash
sudo systemctl daemon-reload
sudo systemctl enable ddns-updater.timer
sudo systemctl start ddns-updater.timer
```
