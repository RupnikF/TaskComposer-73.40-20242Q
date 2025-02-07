# Commands

### Bash
Runs a command through bash. Pipes and redirect work.
```yaml
steps:
  - run-bash:
      service: "ubuntu_service"
      task: "bash"
      input:
        cmd: "compgen -c" # lists all available commands
  - replace:
      service: "ubuntu_service"
      task: "bash"
      input:
        cmd: "echo {{hola}}"
        hola: "ajajaja" # replaces hola inside the cmd. Works with 
```

Output
```json
{
  "stderr": "",
  "stdout": "ajajaja"
}
```

### Eval
Runs an eval on an expression similar to the eval() in JS, uses the bc command.
```yaml
steps:
  - run-eval:
    service: "ubuntu_service"
    task: "eval"
    input:
      exp: "{{A}} + 2"
      A: "23"
```

Outputs
```json
    {
       "result": "25" 
    }
```


### Error format
```json
{
  "error": {
    "msg": "Error message"
  }
}
```