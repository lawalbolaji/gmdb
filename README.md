# Go Memory Database Server (GMDB)

GMDB is an memory database server that is compatible with redis clients. It was developed as a way to learn the internals of Redis and for general education, and is not intended to be used as a production server.

![gmdb demo](https://github.com/lawalbolaji/gmdb/assets/22568024/1a91fac9-412b-491c-8fbc-ae5a1d8e2c55)


## ✨ Features

- GMDB is compatible with standard redis clients e.g cli, and sdks for Node, python, etc.

- It supports SET, GET, HSET, HGET, HGETALL, PING and Atomic Transactions via MULTI command.

## ⚡️ Installation

You need to have go (version 1.22.0 or later) installed, see [how to install go](https://go.dev/doc/install)

Run:

```sh
go install https://github.com/lawalbolaji/gmdb@latest
```

## Local Development

If you want to open the guts of gmdb and experiment, you'll need to have go (version 1.22.0 or later) installed, see [how to install go](https://go.dev/doc/install)

Clone this repo:

```sh
git clone https://github.com/lawalbolaji/gmdb
```

Hack away!!
