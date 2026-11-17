# Inventory API

A RESTful inventory management backend built with [Go](https://go.dev/), [Gin](https://github.com/gin-gonic/gin) and [GORM](https://gorm.io/) on PostgreSQL. Items have a UUID, name, stock and price. The API supports CRUD, pagination, sorting, filtering and rate limiting.

> Status: **skeleton**. Search the code for `TODO(n)` to find what to build next (`grep -rn "TODO(" .`).

## Project layout

```
main.go              entry point (go run .)
config/              DSN + port
models/item.go       Item struct
database/            GORM connection, migration, seed data
handlers/item.go     one handler per endpoint
middleware/          rate limiter
routes/routes.go     Gin router + route table
```

## Installation

Requires Go 1.23+ and PostgreSQL.

```bash
git clone <this-repo> && cd inventory-api
go mod tidy
```

### Database setup

```bash
createdb inventory                      # or: psql -U postgres -c "CREATE DATABASE inventory;"
export DATABASE_DSN="host=localhost user=postgres password=your_password dbname=inventory port=5432 sslmode=disable"
```

If `DATABASE_DSN` is not set, the default in `config/config.go` is used. The table is created by GORM `AutoMigrate` and seeded with 20 items on first run.

Check the data: `psql inventory -c "SELECT * FROM items;"`

## Launch

```bash
go run .
```

Server: http://localhost:8080

## Usage

| Method | Path             | Description                 |
|--------|------------------|-----------------------------|
| GET    | `/inventory`     | List items (paginate/sort/filter) |
| GET    | `/inventory/:id` | Get one item                |
| POST   | `/inventory`     | Create an item              |
| PUT    | `/inventory/:id` | Update an item              |
| DELETE | `/inventory/:id` | Delete an item              |

Query params for `GET /inventory` (TODO(6)): `limit`, `offset`, `sort_by` (`name|stock|price`), `order` (`asc|desc`), `name`, `min_stock`.

```bash
curl "localhost:8080/inventory?limit=5&offset=0&sort_by=price&order=desc&min_stock=10"
curl -X POST localhost:8080/inventory -H 'Content-Type: application/json' -d '{"name":"Dock","stock":4,"price":59.99}'
```

Rate limit: 1 request/second with a burst of 5 (HTTP 429 when exceeded). TODO(5)

## TODO roadmap

1. Postgres + GORM connection (`config`, `database`)
2. `Item` struct tags + validation (`models`)
3. Seed + verify (`database/seed.go`)
4. Five CRUD handlers with correct status codes (`handlers`)
5. Rate limiting middleware; run `go get golang.org/x/time/rate` (`middleware`)
6. Pagination, sorting, filtering on `GET /inventory`
7. Test edge cases (bad input, missing IDs, 429) with Postman/curl; finish this README
8. Optional extras: Swagger (`swaggo/swag`) and in-memory caching
