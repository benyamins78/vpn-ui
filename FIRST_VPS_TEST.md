# First disposable VPS test

Use a fresh Ubuntu 24.04 or Debian 12 amd64 VPS and keep provider console
access available. The current workspace cannot publish a fork because no GitHub
authentication is configured, so substitute the real published fork in
`VPN_UI_REPO` after pushing it.

```bash
export VPN_UI_REPO='OWNER/REPOSITORY'
bash <(curl -fsSL "https://raw.githubusercontent.com/${VPN_UI_REPO}/codex/unified-control-plane/deploy.sh")
```

1. Provision the clean VPS and connect as root (or confirm `sudo` works).
2. Run the installer above. For an unattended BBR opt-out, add `--no-bbr`.
3. Confirm the installer prints the detected architecture, fork/release, and
   BBR result. Unsupported BBR must not abort installation.
4. Check the service: `systemctl status vpn-ui --no-pager`.
5. Open the printed panel URL and sign in with the randomized credentials.
6. Create one test inbound and client.
7. Connect from a real Windows, macOS, Linux, or Android device.
8. Confirm the Clients page Online count changes.
9. Generate traffic, then inspect the client's 1h/24h history endpoint/page.
10. Disconnect and confirm the client becomes Offline after the next tick.
11. Reboot: `systemctl reboot`.
12. Confirm `systemctl is-active vpn-ui` reports `active` after boot.
13. Confirm BBR persistence:
    `sysctl net.ipv4.tcp_congestion_control` and
    `sysctl net.core.default_qdisc`.
14. Confirm the client, inbounds, credentials, totals, and history remain.
15. Inspect `journalctl -u vpn-ui -b --no-pager -n 200` for errors.

## BBR panel checks

Open the dashboard's TCP congestion control card. Verify current algorithm,
qdisc, kernel, availability, and persistence. On a supported kernel use Enable
BBR, verify `bbr`/`fq`, then use Revert vpn-ui BBR and confirm only
`/etc/sysctl.d/99-vpn-ui-bbr.conf` is removed and the previous values return.

## Idempotent upgrade and uninstall

Run the same installer again. It must detect an update, back up the database,
preserve credentials, and preserve existing BBR settings unless `--enable-bbr`
is explicitly supplied. Verify the service restarts and all data remains.

Use the installed binary's documented uninstall command only after taking a
backup, then verify unrelated systemd units, sysctl files, and networking remain
untouched. Reinstall with the same command and verify the panel starts cleanly.

## Manual verification commands

```bash
systemctl status vpn-ui --no-pager
journalctl -u vpn-ui -e --no-pager
sysctl net.ipv4.tcp_congestion_control
sysctl net.ipv4.tcp_available_congestion_control
sysctl net.core.default_qdisc
ls -l /etc/sysctl.d/99-vpn-ui-bbr.conf
```
