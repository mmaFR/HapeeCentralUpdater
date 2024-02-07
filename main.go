package main

import (
	idb "HapeeCentralUpdater/internal_db"
	"fmt"
	"log"
	"net/http"
)

const mapPath string = "path.map"
const aclPath string = "src.acl"

func main() {
	var err error
	var internalDatabases map[string]*idb.InternalDB = make(map[string]*idb.InternalDB)
	var iDB *idb.InternalDB
	var mux *http.ServeMux = http.NewServeMux()

	if iDB, err = idb.NewInternalDb(idb.ValueDb, aclPath); err != nil {
		log.Fatalln(err)
	} else {
		internalDatabases[aclPath] = iDB
		if err = iDB.Load(); err != nil {
			log.Fatalln(err.Error())
		}
	}

	if iDB, err = idb.NewInternalDb(idb.KeyValueDb, mapPath); err != nil {
		log.Fatalln(err)
	} else {
		internalDatabases[mapPath] = iDB
		if err = iDB.Load(); err != nil {
			log.Fatalln(err.Error())
		}
	}

	mux.HandleFunc("/", http.NotFound)
	for _, s := range []string{mapPath, aclPath} {
		mux.HandleFunc(fmt.Sprintf("/%s/add", s), internalDatabases[s].HttpAddHandler)
		mux.HandleFunc(fmt.Sprintf("/%s/del", s), internalDatabases[s].HttpDelHandler)
		mux.HandleFunc(fmt.Sprintf("/%s/get", s), internalDatabases[s].HttpGetHandler)
	}

	if err = http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
	}
}
