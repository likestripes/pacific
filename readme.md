## Pacific

A golang lib to integrate all the clouds.

### Warning

*This is v.01 -- it has no tests, performance is probably awful and it's not proven safe in production anywhere.*  But maybe it'll scratch an itch?

### WTF

`Pacific` is an opinionated query wrapper that makes Google's Go Datastore interchangeable with Postgres (and :calendar: others!).

### Install / Import

`go get -u github.com/likestripes/pacific`

```go
import (
	"github.com/likestripes/pacific"
)
```

Google AppEngine: `goapp serve` works out of the box (they include the buildtag for you)

Postgres: `go run -tags 'postgres' main.go` -- details below.

### Postgres Options

`pacific_pg_user=foo pacific_pg_dbname=bar go run -tags 'postgres' main.go`

- pacific_pg_user
- pacific_pg_password
- pacific_pg_dbname
- pacific_pg_host - defaults to localhost
- pacific_pg_port - defaults to 5432
- pacific_pg_sslmode - defaults to disabled

- pacific_log - log verbosely
- pacific_migrate - automagically migrate models

### Query

Queries are a struct:

```go
type Query struct {
	Kind       string
	Context    Context
	Offset     int
	Limit      int
	KeyString  string
	KeyInt     int64
	Ancestors  []Ancestor
	Order      string
	Filters    map[string]string
	PrimaryKey string
}
```

that can be used minimally to get a thing:

```go

context := pacific.NewContext(r) //r = *http.Request

query := pacific.Query{
  Context:   context,
  Kind:      "Thing",
  KeyString: thing_id, //if you prefer Int, you'd use KeyInt instead
}

var thing Thing
err := query.Get(&thing)

return thing
```

or with friends:


```go
context := pacific.NewContext(r) //r = *http.Request

query := pacific.Query{
  Context: context,
  Kind:    "Friend",
  Limit:   10,
}

var ten_friends []Friend
err := query.GetAll(&ten_friends)

if err != nil {
  context.Errorf(err.Error())
}

return ten_friends
```

### Queries implement:

```go
query.Put(entry interface{}) error
query.Get(result interface{}) error
query.GetAll(results interface{}) error
query.Delete(result interface{}) error
```

### Joining your `Ancestor`

Ancestors allow you to nest your models (in AppEngine) or join your tables (in Postgres)

```go
type Ancestor struct {
	Context    Context
	Kind       string
	KeyString  string
	KeyInt     int64
	Parent     *datastore.Key
	PrimaryKey string
}
```

Pass an order-sensitive array of `Ancestor`s in a `Query` and it will recurse from first to last. In AppEngine, this is simply chaining your keys; in Postgres, it's more like a join.


#### TODO (Meat & Potatoes)
- [ ] extensible file structure ("postgres" "psqlite" etc subpacks?)
- [ ] logging
- [ ] documentation!
- [ ] tests!
- [ ] benchmarking

#### Feature Requests
- [ ] better indexing for PG
- [ ] urlfetch? s3? should this extend beyond datastores?
- [ ] sqlite
- [ ] mongo
- [ ] dynamodb

- [x] Contributors welcome!
