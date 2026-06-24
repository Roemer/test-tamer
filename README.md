# Test Tamer

## Requirements

Postgres 18

## Configuration

`TEST_TAMER_SUBTITLE`: Defines a subtitle to be shown on the dashboard. Set to "" to hide.

## Development

Make sure to initialize the vendor scripts:

```
go run "./build" --target "vendor:init"
```

Run the server with fake data:

```
go run . --fake
```
