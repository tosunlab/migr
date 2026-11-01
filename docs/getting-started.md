---
title: Getting started
---

# Getting started

This page installs migrasyon, writes a first migration in SQL, and runs it against a SQLite file, so you
need no database server. Then it shows what changes for PostgreSQL or MySQL, and where to go next.

## Install

Pick one. Each installs the same binary.

```console
$ brew install tosunlab/tap/migrasyon          # macOS or Linux, with shell completions
$ npm install --save-dev @migrasyon/cli       # in a JavaScript project, then `npx migrasyon`
$ docker pull ghcr.io/tosunlab/migrasyon       # for pipelines; see the zero-downtime guide's CI job
```

Or download an archive for your platform from
[GitHub Releases](https://github.com/tosunlab/migrasyon/releases). The
[README](../README.md#installing) shows how to verify its signature.

```console
$ migrasyon version
```

## Start a project

```console
$ mkdir app && cd app
$ migrasyon init
created migrasyon.yaml
created migrations
next: `migrasyon create add_users`, and set DATABASE_URL before `migrasyon up`
```

`migrasyon init` writes a `migrasyon.yaml` for PostgreSQL. For this walkthrough, point it at a SQLite file
instead. Change the two lines under `databases.default`:

```yaml
databases:
  default:
    driver: sqlite
    dsn: ./app.db
```

`migrasyon config` prints the whole resolved configuration, every default included, with secrets
redacted.

## Write a migration

```console
$ migrasyon create add_users
created migrations/20261023T171330_add_users.sql
```

The file name's timestamp is the migration's version, and versions run in order. Replace the file's
template with:

```sql
-- migrasyon:up
CREATE TABLE users (
  id    INTEGER PRIMARY KEY,
  email TEXT NOT NULL UNIQUE
);

-- migrasyon:down
DROP TABLE users;
```

## See what will run, then run it

`migrasyon plan` compiles every pending migration and prints what it would send, without sending anything:

```console
$ migrasyon plan
plan — database "default", up, batch 1

this database has no migration history table yet; migrasyon up would create it

  1  20261023T171330_add_users.sql
       1  raw  native
          CREATE TABLE users (
            id    INTEGER PRIMARY KEY,
            email TEXT NOT NULL UNIQUE
          );

1 migration, 1 operation, 1 statement. nothing was executed and nothing was recorded.
```

```console
$ migrasyon up
up — database "default", batch 1

  1  20261023T171330_add_users.sql  1 statement  162µs

applied 1 migration in 162µs

$ migrasyon status
status — database "default"

  applied  20261023T171330_add_users.sql  up in batch 1, 2026-10-23 17:13:36 UTC

1 migration: 1 applied
```

`migrasyon history` shows every run, and `migrasyon down` rolls the newest batch back:

```console
$ migrasyon down
down — database "default", batch 2

  1  20261023T171330_add_users.sql  1 statement  204µs

rolled back 1 migration in 204µs
```

Each migration runs in one transaction where the engine allows it, and migrasyon records it in a history
table (`migrasyon_migrations`) in that same transaction. Once applied, a migration's file is fixed:
editing it is **checksum drift**, and `migrasyon up` refuses (exit 4) rather than guess what you meant.

## Use your real database

Point `dsn` at it, from the environment:

```yaml
databases:
  default:
    driver: postgres          # or mysql, for MySQL and MariaDB
    dsn: ${DATABASE_URL}
```

Nothing else changes. Each driver's page lists what it supports, what each operation costs a live
database, and what it refuses: [PostgreSQL](drivers/postgres.md),
[MySQL and MariaDB](drivers/mysql.md), [SQLite](drivers/sqlite.md).

## Choose a language

SQL is the quickest start, but a `.sql` migration is raw engine SQL. migrasyon can't read it, so it can't
derive its down, judge its lock cost, or run it on another engine. A **Go** or **TypeScript**
migration describes the change instead, so it runs on every driver, its down can be derived, and
every operation's cost is known before it runs.

- [Go](guides/go.md)
- [TypeScript](guides/typescript.md)
- [SQL](guides/sql.md)

One directory can hold more than one language: each file runs in the language its extension names.
`migrations.language` only decides what `migrasyon create` writes.

## Exit codes

Every command exits with a code a script can branch on:

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | the world isn't ready: an unreachable database, a dirty run, a missing migration |
| 2 | the argument is wrong: a bad flag, a refused migration, a validation problem |
| 3 | still waiting for a lock after every retry |
| 4 | an applied migration's file changed |
| 5 | `--require-safe` found an operation above `safety.max_lock_impact` |
| 6 | a shadow-table migration is waiting for its cutover to be confirmed |

Every failure says what failed and why, then what to do about it on a line of its own:

```console
$ migrasyon down
migrasyon: 20261023T171330_add_users.sql: this migration declares that it cannot be undone: the rows are gone
  next: roll back only what is newer with migrasyon down --to <its version>; only a backup taken before it ran reaches further
```

`--json` turns every command's output, and every failure, into one JSON document per stream. A
failure's document carries the same remedy as `next`.

## Next

- Gate CI on what a migration costs a live database, change a column across deploys, and run large
  changes through a shadow table: the [zero-downtime guide](zero-downtime.md).
- Draw the dependency DAG and the schema each migration builds, as text, Mermaid or Graphviz:
  [visualizing migrations](visualizing.md).
- Four complete projects, each with tests and a CI workflow to copy:
  [Go on PostgreSQL](../examples/go-postgres/README.md),
  [TypeScript on MySQL](../examples/ts-mysql/README.md),
  [all three languages on SQLite](../examples/mixed-sqlite/README.md), and a
  [zero-downtime column change](../examples/zero-downtime/README.md).
- Write an SDK of your own against the [operation IR](ir.md) and the
  [runtime protocol](protocol.md).
