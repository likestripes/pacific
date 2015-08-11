// +build postgres

package pacific

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
)

type Context struct {
	Connection gorm.DB
	Tables     map[string]bool
}

func Main() {
	//no-op
}

func IsDevAppServer() bool {
	return true
}

func SupportsWS() bool {
	return true
}

func log_from_env() bool {
	logging := os.Getenv("pacific_log")
	if logging == "details" {
		return true
	}
	return false
}

func migration_from_env() bool {
	migration := os.Getenv("pacific_migrate")
	if migration == "yes" {
		return true
	}
	return false
}

func db_from_env() (user, password, dbname, host, port, sslmode string) {
	return os.Getenv("pacific_pg_user"), os.Getenv("pacific_pg_password"), os.Getenv("pacific_pg_dbname"), os.Getenv("pacific_pg_host"), os.Getenv("pacific_pg_port"), os.Getenv("pacific_pg_sslmode")
}

func connection_string() string {
	user, password, dbname, host, port, sslmode := db_from_env()

	var password_str string
	if password != "" {
		password_str = fmt.Sprintf("password=%s", password)
	}

	if sslmode == "" {
		sslmode = "disable"
	}

	if host == "" {
		host = "localhost"
	}

	if port == "" {
		port = "5432"
	}

	return fmt.Sprintf("user=%s %s dbname=%s host=%s port=%s sslmode=%s", user, password_str, dbname, host, port, sslmode)
}

func NewContext(r *http.Request) Context {
	DB, _ := gorm.Open("postgres", connection_string())
	if r != nil {
		DB.DB()
		if log_from_env() {
			DB.LogMode(true)
		}
	}
	return Context{DB, make(map[string]bool)}
}

func (context Context) Infof(obj interface{}) {
	log.Print(obj)
}

func (context Context) Errorf(obj interface{}) {
	log.Print(obj)
}

type Ancestor struct {
	Kind       string
	KeyString  string
	KeyInt     int64
	PrimaryKey string
	Key        interface{}
	Context    Context
}

type Query struct {
	Kind       string
	tableName  string
	PrimaryKey string
	migrated   bool
	Context    Context
	Table      *gorm.DB
	Offset     int
	Limit      int
	KeyString  string
	KeyInt     int64
	Key        interface{}
	Ancestors  []Ancestor
	Order      string
	Filters    map[string]string
}

func (query Query) Delete() (err error) {

	query.key()

	if query.Key != nil {
		q := query.table(nil).Where(query.PrimaryKey+" = ?", query.Key)

		for _, ancestor := range query.Ancestors {
			filter_by_str := ancestor.primaryKey() + " = ?"
			q = q.Where(filter_by_str, ancestor.key())
		}

		q.Delete(nil)
	}

	if err != nil {
		log.Fatal(err.Error())
	}
	return err
}

func (query Query) Get(dst interface{}) (err error) {
	query.key()
	if query.Key != nil {
		q := query.table(dst).Where(query.PrimaryKey+" = ?", query.Key)

		for _, ancestor := range query.Ancestors {
			filter_by_str := ancestor.primaryKey() + " = ?"
			q = q.Where(filter_by_str, ancestor.key())
		}

		q.First(dst)

	}
	if err != nil {
		log.Fatal(err.Error())
	}
	return err
}

func (query Query) GetAll(results interface{}) (err error) {

	query.createQuery().Find(results)

	if err != nil {
		log.Fatal(err.Error())
	}
	return err
}

func (query Query) Put(entry interface{}) (err error) {

	db := query.table(entry)
	query.key()
	log.Print("PUT of kind: " + query.Kind)
	var count int
	q := db.Where(query.PrimaryKey+" = ?", query.Key)

	for _, ancestor := range query.Ancestors {
		filter_by_str := ancestor.primaryKey() + " = ?"
		q = q.Where(filter_by_str, ancestor.key())
	}

	q.Count(&count)

	if count == 0 {
		db.Save(entry)
	} else {
		q.Updates(entry)
	}

	return err
}

func (query Query) createQuery() (q *gorm.DB) {

	query.tableify()
	q = query.Context.Connection.Table(query.tableName)

	for _, ancestor := range query.Ancestors {
		filter_by_str := ancestor.primaryKey() + " = ?"
		q = q.Where(filter_by_str, ancestor.key())
	}

	for filter_by, value := range query.Filters {
		filter_by_str := filter_by + " = ?"
		q = q.Where(filter_by_str, value)
	}

	if query.Limit != 0 {
		q = q.Limit(query.Limit)
	}

	if query.Offset != 0 {
		q = q.Offset(query.Offset)
	}

	if query.Order != "" {
		direction, column := order_by(query.Order)
		order_str := column + " " + direction
		q = q.Order(order_str)
	}

	return q
}

func order_by(order string) (direction, column string) {

	if order[0:1] == "-" {
		direction = "desc"
		column = order[1:]
	} else {
		direction = "asc"
		column = order
	}

	return
}

func AutoMigrate(context Context, kind string, primary_key string, dst interface{}){
	tableName := strings.ToLower(kind) + "s"
	table := context.Connection.Table(tableName)
	table.AutoMigrate(dst)
	compositeIndex(table, kind, primary_key, dst, []string{})
}


func (query *Query) table(dst interface{}) *gorm.DB { //TODO: this should use Context.Tables to memoize

	query.tableify()
	query.key()

	query.Table = query.Context.Connection.Table(query.tableName)

	if migrated, ok := query.Context.Tables[query.tableName]; migration_from_env() && dst != nil && !query.migrated && (!ok || !migrated) {
		log.Print("running automigrate and idx creation")
		query.Table.AutoMigrate(dst)
		query.indexPrimaryKey(dst)
		query.migrated = true
		query.Context.Tables[query.tableName] = true
	}

	return query.Table
}

func (query *Query) tableify() string {
	query.primaryKey()
	query.tableName = strings.ToLower(query.Kind) + "s"
	return query.tableName
}

func (query *Query) key() interface{} {

	query.primaryKey()

	if query.KeyString != "" {
		query.Key = query.KeyString
	}

	if query.KeyInt != 0 {
		query.Key = query.KeyInt
	}

	return query.Key
}

func (query *Query) primaryKey() string {
	if query.PrimaryKey == "" {
		query.PrimaryKey = strings.ToLower(query.Kind) + "_id"
	}
	return query.PrimaryKey
}

func (ancestor *Ancestor) primaryKey() string {
	if ancestor.PrimaryKey == "" {
		ancestor.PrimaryKey = strings.ToLower(ancestor.Kind) + "_id"
	}
	return ancestor.PrimaryKey
}

func (ancestor *Ancestor) key() interface{} {

	ancestor.primaryKey()

	if ancestor.KeyString != "" {
		ancestor.Key = ancestor.KeyString
	}

	if ancestor.KeyInt != 0 {
		ancestor.Key = ancestor.KeyInt
	}

	return ancestor.Key
}

func (query *Query) indexPrimaryKey(dst interface{}) {
	query.primaryKey()

	if query.PrimaryKey != "" {
		parents := []string{}

		if len(query.Ancestors) > 0 {
			for _, ancestor := range query.Ancestors {
				parents = append(parents, ancestor.PrimaryKey)
			}
		}

		compositeIndex(query.Table, query.Kind, query.PrimaryKey, dst, parents)
	}
	return
}

func compositeIndex(table *gorm.DB, kind string, primary_key string, dst interface{}, parents []string) {

	index_name := "idx_"+ kind +"_"+ primary_key
	indexes := []string{primary_key}

	if len(parents) == 0 {
		parents = []string{}
		st := reflect.TypeOf(dst).Elem()
		for i := 0; i < st.NumField(); i++ {
			field := st.Field(i)
			pacific_parent := field.Tag.Get("pacific_parent")
			if pacific_parent != "" {
				parents = append(parents, pacific_parent)
			}
		}
	}

	for _, parent := range parents {
		if parent != "" {
			index_name = index_name + "_" + parent
			indexes = append(indexes, parent)
		}
	}

	if len(indexes) > 0 {
		table.AddUniqueIndex(index_name, indexes...) //TODO: if not exists?
	}
}
