# ForgeAI
daily and wikely task

## Backend Structure

```text
Backend/
|
|-- cmd/
|   |-- agent/
|   |   `-- main.go
|   `-- server/
|       `-- main.go
|
|-- internal/
|   |-- modules/
|   |   |-- agent/
|   |   |   |-- handler.go
|   |   |   |-- service.go
|   |   |   |-- repository.go
|   |   |   |-- model.go
|   |   |   `-- route.go
|   |   |-- conversation/
|   |   |   |-- handler.go
|   |   |   |-- service.go
|   |   |   |-- repository.go
|   |   |   `-- model.go
|   |   |-- llm/
|   |   |   |-- handler.go
|   |   |   |-- service.go
|   |   |   |-- repository.go
|   |   |   `-- model.go
|   |   |-- tool/
|   |   |   |-- handler.go
|   |   |   |-- service.go
|   |   |   |-- repository.go
|   |   |   `-- model.go
|   |   |-- terminal/
|   |   |   |-- handler.go
|   |   |   |-- service.go
|   |   |   |-- repository.go
|   |   |   `-- model.go
|   |   |-- workspace/
|   |   |   |-- handler.go
|   |   |   |-- service.go
|   |   |   |-- repository.go
|   |   |   `-- model.go
|   |   |-- permission/
|   |   |   |-- handler.go
|   |   |   |-- service.go
|   |   |   `-- model.go
|   |   `-- git/
|   |       |-- handler.go
|   |       |-- service.go
|   |       `-- repository.go
|   |-- middleware/
|   |   |-- auth.go
|   |   `-- recovery.go
|   |-- config/
|   |   `-- config.go
|   `-- database/
|       `-- postgres.go
|
|-- migrations/
|-- prompts/
|-- configs/
`-- go.mod
```
