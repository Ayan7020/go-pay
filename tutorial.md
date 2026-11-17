### Installing Migration 
```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Creating migration
```
migrate create -ext sql -dir {{dir-to-migration}} -seq {{migration-name}}
```