# Test Tamer

## Requirements

Postgres 18

## Development

Make sure to initialize the vendor scripts:

```
go run "./build" --target "vendor:init"
```

Run the server with fake data:

```
go run . --fake
```
