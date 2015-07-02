// +build appengine

package pacific

import (
	"appengine"
	"appengine/datastore"
	"net/http"
)

type MockConnection struct{}

func (conn MockConnection) Close() {
	return
}

type Context struct {
	appengine.Context
	Connection MockConnection
}

func NewContext(r *http.Request) Context {
	context := appengine.NewContext(r)
	return Context{context, MockConnection{}}
}

func Main() {
	//no-op
}

func IsDevAppServer() bool {
	return appengine.IsDevAppServer()
}

func SupportsWS() bool {
	return false
}

type Ancestor struct {
	Context    Context
	Kind       string
	KeyString  string
	KeyInt     int64
	Parent     *datastore.Key
	PrimaryKey string
}

func (ancestor Ancestor) key() *datastore.Key {
	return datastore.NewKey(ancestor.Context, ancestor.Kind, ancestor.KeyString, ancestor.KeyInt, ancestor.Parent)
}

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

func (query Query) Delete() error {
	key := query.key()
	return query.deleteByKey(key)
}

func (query Query) Get(result interface{}) error {
	key := query.key()
	return datastore.Get(query.Context, key, result)
}

func (query Query) GetAll(results interface{}) error {
	_, err := query.createQuery().GetAll(query.Context, results)
	return err
}

func (query Query) Put(entry interface{}) error {
	key := query.key()
	_, err := datastore.Put(query.Context, key, entry)
	return err
}

func (query Query) key() *datastore.Key {
	ancestor_key := query.ancestorKey()
	return datastore.NewKey(query.Context, query.Kind, query.KeyString, query.KeyInt, ancestor_key)
}

func (query Query) ancestorKey() (parent *datastore.Key) {
	if len(query.Ancestors) > 0 {
		for _, ancestor := range query.Ancestors {
			ancestor.Context = query.Context
			ancestor.Parent = parent
			parent = ancestor.key()
		}
		return parent
	}
	return nil
}

func (query Query) createQuery() (q *datastore.Query) {

	q = datastore.NewQuery(query.Kind)

	if len(query.Ancestors) > 0 {
		ancestor_key := query.ancestorKey()
		q = q.Ancestor(ancestor_key)
	}

	for filter_by, value := range query.Filters {
		q = q.Filter(filter_by, value)
	}

	if query.Limit != 0 {
		q = q.Limit(query.Limit)
	}

	if query.Offset != 0 {
		q = q.Offset(query.Offset)
	}

	if query.Order != "" {
		q = q.Order(query.Order)
	}

	return q
}

func (query Query) deleteByKey(key *datastore.Key) error {
	return datastore.Delete(query.Context, key)
}
