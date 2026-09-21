# ForgeAI
daily and wikely task

## Backend Structure

```text
Backend/
|-- cmd/
|   `-- agent/
|       `-- main.go
|-- internal/
|   |-- modules/
|   |   |-- agent/        (handler.go, service.go, repository.go, model.go, route.go)
|   |   |-- conversation/ (handler.go, service.go, repository.go, model.go)
|   |   |-- llm/          (service.go, repository.go, model.go)
|   |   |-- tool/         (service.go, registry.go, model.go)
|   |   |-- workspace/    (service.go, repository.go, model.go)
|   |   |-- permission/   (service.go, repository.go, model.go)
|   |   `-- git/          (service.go, repository.go, model.go)
|   |-- shared/
|   |   |-- executor/     (contract and result types)
|   |   |   |-- runtime/  (runner.go, runner_test.go)
|   |   |   |-- shell/    (executor.go, bash.go, cmd.go, powershell.go, zsh.go)
|   |   |   `-- factory/  (factory.go, factory_test.go)
|   |   |-- filesystem/   (filesystem.go)
|   |   |-- errors/       (errors.go)
|   |   |-- logger/       (logger.go)
|   |   `-- types/        (common.go)
|   |-- middleware/       (recovery.go, logging.go)
|   |-- config/           (config.go)
|   `-- database/         (postgres.go)
|-- prompts/
|-- configs/
|-- migrations/
|-- go.mod
`-- go.sum
```
