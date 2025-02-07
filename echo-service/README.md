# Commands

### Echo
Receives a message and returns
```yaml
steps:
  - run-echo:
      service: "echo_service"
      task: "echo"
      input:
        msg: "message" # lists all available commands
```

Output
```json
{
  "msg": "message"
}
```