# hass-tray

A simple Go application to display basic state details (via a tray icon) from a Home Assistant instance.

- Ultra simple configuration
  - YAML configuration file
  - Environment variables
  - Dotenv file
- Resistant to poor network conditions
- Lightweight, simple icons, simple UI options

## Wishlist

- [ ] Handle disconnections, reconnect automatically
- [ ] Event listening with watchdog
- [ ] Rotating System Logs
- [ ] Tray Icon Context Menu (Quit, Refresh, Open Log, Open Home Assistant)
- [ ] State Change Icons
  - Icons that better show recent state changes (e.g. Orange for recently closed, Purple for bad connection)