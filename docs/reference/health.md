# Health

Check server status and availability.

## Bridge Mode

```bash
curl http://localhost:9867/health
# CLI Alternative
pinchtab health
# Response
{
  "status": "ok",
  "tabs": 1
}
```

Notes:

- returns tab count for the attached browser
- in error cases returns `503` with `status: "error"`

## Server Mode (Dashboard)

```bash
curl http://localhost:9867/health
# Response
{
  "status": "ok",
  "mode": "dashboard",
  "version": "0.8.0",
  "uptime": 12345,
  "profiles": 1,
  "instances": 1,
  "defaultInstance": {
    "id": "inst_abc12345",
    "status": "running"
  },
  "agents": 0,
  "restartRequired": false
}
```

| Field | Description |
|-------|-------------|
| `status` | `ok` when server is healthy |
| `mode` | Always `dashboard` in server mode |
| `version` | PinchTab version |
| `uptime` | Milliseconds since server start |
| `profiles` | Number of configured profiles |
| `instances` | Number of running browser instances |
| `defaultInstance` | First managed instance info (if any) |
| `defaultInstance.id` | Instance ID |
| `defaultInstance.status` | `starting`, `running`, `stopping`, `stopped`, `error` |
| `agents` | Number of connected agents |
| `restartRequired` | True if config changes need restart |
| `restartReasons` | List of reasons (when `restartRequired` is true) |

Notes:

- `defaultInstance` is present when at least one instance is running
- use `defaultInstance.status == "running"` to check Chrome is ready
- strategies like `always-on` launch an instance at startup

## Related Pages

- [Tabs](./tabs.md)
- [Navigate](./navigate.md)
- [Strategies](./strategies.md)
