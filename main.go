package pacific

import (
	// appengine classic requirements:
	// _ "appengine"
	// _ "appengine/datastore"

	// postgres requirements:
	// _ "github.com/jinzhu/gorm"
	// _ "github.com/lib/pq"

	// appengine managed-vm requirements:
	_ "golang.org/x/net/context"
	_ "google.golang.org/appengine"
	_ "google.golang.org/appengine/datastore"
	_ "google.golang.org/appengine/log"
	_ "google.golang.org/appengine/urlfetch"
)
