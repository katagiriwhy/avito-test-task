# Приветствую команду Авито!

Моя реализация тестового задания. Ниже представлены инструкции по сборке и запуску проекта, а также ответы на дополнительные задания.

---

##  Сборка и запуск проекта

### Сборка
```bash
make build
```

### Запуск
```bash
make up
```
---

## Дополнительные задания

### 1. Массовая деактивация пользователей команды

Реализован эндпоинт, позволяющий деактивировать всех пользователей по названию команды, а также безопасно переназначать их открытые Pull Request’ы.

**Endpoint:**

```bash
POST localhost:8080/users/deactivate?team_name=<team_name>
```

### 2. Статистика по назначению пользователей на pullRequests

```bash
GET localhost:8080/stats/reviewAssignments
```

### 3. Конфигурация линтера (`.golangci.yml`)

```yaml
run:
  timeout: 3m
  issues-exit-code: 1

linters:
  enable:
    - govet
    - staticcheck
    - errcheck
    - revive
    - gofmt
    - goimports

linters-settings:
  revive:
    confidence: 0.8
    severity: warning
    rules:
      - name: exported
        arguments:
          - allow-leading-underscore
      - name: var-naming
      - name: package-comments

issues:
  exclude-use-default: false
  max-issues-per-linter: 50
  max-same-issues: 3
```