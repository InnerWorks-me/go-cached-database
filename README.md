# go-cached-database

This library provides an opinionated Database adaptor for PostgreSQL 
that supports Redis as a read-through + write-through cache.

## Tech Stack

The following dependencies are used:

- [sqlc](https://github.com/sqlc-dev/sqlc) - Generate type-safe code from SQL.
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit for Go.
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database Migrations Opeartor.

## Getting Started

### Pre-reqs
- You must locally [install sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) to be able to generate Go code from your queries:
```
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Step 1 - Write Migrations

Start by defining your database schema in the form of migrations. [golang-migrate](https://github.com/golang-migrate/migrate) defines each migration as a set of two files:
```
{version}_{title}.up.sql
{version}_{title}.down.sql
```
The `title` of each migration is unused, and is only for readability.
Versions of migrations may be represented as any 64 bit unsigned integer. All migrations are applied upward in order of increasing version number, and downward by decreasing version number.

The recommended versioning schema is to use a timestamp at an appropriate resolution:

```
1500360784_initialize_schema.down.sql
1500360784_initialize_schema.up.sql
1500445949_add_table.down.sql
1500445949_add_table.up.sql
...
```

All migrations must be placed inside a dedicated directory which must be shipped together with the compiled binary.

You can find example migrations at [migrations_test](./migrations_test).

### Step 2 - Write Queries

As we're using `sqlc`, our goal is to write the SQL queries that we intend to use in our program, and `sqlc` will take care of generating the necessary Go code.

Queries must be placed into a single file, normally called `queries.sql`.
All queries must be [annotated](https://docs.sqlc.dev/en/latest/reference/query-annotations.html) with a `name` and an `execution context`. Here's an example:

```sql
-- name: GetAuthor :one
SELECT * FROM authors
WHERE id = $1 LIMIT 1;
```

In this example, `name` is set to `GetAuthor` and the `execution context` is set to `one`. That will tell `sqlc` to generate a query called `GetAuthor` and that it should return `one` single row.

You can find more examples at [query.sql](./queries_test/query.sql).

### Step 3 - Binding it all Together - Generating Code
Now that we have both queries and migrations defined, we must use `sqlc` to generate the equivalent Go code. Install `sqlc` if you haven't by now.

`sqlc` requires relevant context to be able to generate code. This context is provided through the `sqlc.yaml` configuration file. Here's an example:

```yaml
version: "2"
sql:
  - engine: "postgresql"                # This is a PostgreSQL database.
    schema: "./migrations_test"         # Schema definition. Point it to the migrations folder.
    queries: "./queries_test/query.sql" # Queries. Point it to the queries file.
    gen:
      go:
        package: "queries_test" # Package name into which the generated code will belong.
        out: "./queries_test"   # Output folder. Where the generated files will be saved to.
        sql_package: "pgx/v5"   # A reference to the Go PostgreSQL driver. Leave it as "pgx/v5".
```

After defining the configuration file and saving it as `sqlc.yaml`, run:
```shell
sqlc generate
```
If everything went well, you will find the generated code at `./queries_test`!

### Step 4 - Profit!
Now we're ready to use our newly generated code. Follow the Usage section below for more details.

## Usage

### Instantiate the Adapter
Instantiate the database adapter passing the relevant information. This adapter is thread-safe and should be shared across the code base.
In the following example, we're creating an adapter to the `queries_test.Queries` query set, as defined in the Getting Started section.
```go
adapter, err := NewAdapter[queries_test.Queries](Config[queries_test.Queries]{
    Redis: RedisConfig{
        Endpoint: redisEndpoint,
        Password: "",
        Database: 0,
    },
    Postgres: PostgresConfig{
        Endpoint: postgresEndpoint,
        User:     "postgres",
        Password: "postgres",
        Database: "postgres",
    },
    QueryConstructor: func(db DBTX) *queries_test.Queries {
        return queries_test.New(db)
    },
})
```

### Apply migrations
You'll normally want to apply the database migrations as the program starts:

```go
err := adapter.Migrate(ctx, MigrationConfig{
    MigrationsDir: "./migrations_test",
})
```

### Run simple queries
Create a new Author
```go
newAuthor, err := adapter.Queries.CreateAuthor(ctx, queries_test.CreateAuthorParams{
    Name: "Isaac Newton",
})
fmt.Printf("Inserted new Author with ID: %d\n", newAuthor.ID)
```

Retrieve an Author

```go
authorId := 1
author, err := adapter.Queries.GetAuthor(ctx, authorId)
fmt.Printf("Got Author with ID: %d -> %s\n", author.ID, author.Name)
```

### Run cached queries
In order to cache a query result, you must specify a unique identifier that will be used to index the cached entry. A general good identifier is the object primary key, in the following format:
`table_name:primary_key`. For instance `author:1`:

```go
authorId := 1
cacheKey := fmt.Sprintf("author:%d", authorId)
author, err := WithCache(adapter, cacheKey, func() (queries_test.Author, error) {
    return adapter.Queries.GetAuthor(ctx, newAuthor.ID)
})
```
The example above will attempt to find the author in the cache by its `cacheKey`.
If the author is not present in the cache, 
it will be fetched from the database and then saved to the cache under `cacheKey`, 
effectively implementing a read-through cache pattern.
