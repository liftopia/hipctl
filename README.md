# hipctl - hipache backend configurator

This tool allows you to programmatically add and remove backends from active frontends for hipache.

> Note: As of now, the configurator expects the backends on a particular frontend to use the same port, and all frontends are routed to all backends.

# Installation

```bash
go build
```

# Usage

## Help

```bash
./hipctl
```

## List all details

```bash
./hipctl list
```

## List servers

```bash
./hipctl list servers
```

## Add backend

```bash
./hipctl add <ip>
```

## Remove backend

```bash
./hipctl remove <ip>
```

# License

hipctl is licensed under the MIT License - see the LICENSE file for details

