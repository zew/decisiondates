package gorpx

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/zew/decisiondates/config"
	"github.com/zew/decisiondates/mdl"
	"github.com/zew/gorp"
	"github.com/zew/logx"
	"github.com/zew/util"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var dbmap *gorp.DbMap
var db *sql.DB

func DB() *sql.DB {
	if db == nil {
		DBMap()
	}
	return db
}

func DBMap(dbName ...string) *gorp.DbMap {

	if dbmap != nil && db != nil {
		return dbmap
	}

	sh := config.Config.SQLHosts[util.Env()]
	var err error
	// param docu at https://github.com/go-sql-driver/mysql
	paramsJoined := "?"
	for k, v := range sh.ConnectionParams {
		paramsJoined = fmt.Sprintf("%s%s=%s&", paramsJoined, k, v)
	}

	if len(dbName) > 0 {
		sh.DBName = dbName[0]
	}

	if config.Config.SQLite {
		db, err = sql.Open("sqlite3", "./main.sqlite")
		util.CheckErr(err)
	} else {
		connStr2 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", sh.User, util.EnvVar("SQL_PW"), sh.Host, sh.Port, sh.DBName, paramsJoined)
		logx.Printf("gorp conn: %v", connStr2)
		db, err = sql.Open("mysql", connStr2)
		util.CheckErr(err)
	}

	err = db.Ping()
	util.CheckErr(err)
	logx.Printf("gorp database connection up")

	//
	//
	//
	{
		mp := IndependentDbMapper(db)
		t := mp.AddTable(mdl.Community{})
		t.ColMap("community_key").SetUnique(true)
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			mp.CreateIndex()
		}
	}

	{
		mp := IndependentDbMapper(db)
		t := mp.AddTable(mdl.Pdf{})
		t.SetUniqueTogether("community_key", "pdf_url")
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			mp.CreateIndex()
		}
	}

	{
		mp := IndependentDbMapper(db)
		t := mp.AddTable(mdl.Page{})
		t.SetUniqueTogether("page_url", "page_number")
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			mp.CreateIndex()
		}
	}

	{
		mp := IndependentDbMapper(db)
		t := mp.AddTable(mdl.Decision{})
		// t.ColMap("domain_name").SetUnique(true)
		// t.AddIndex("idx_name_desc", "Btree", []string{"domain_name", "rank_code"})
		t.SetUniqueTogether("community_key", "decision_for_year") //
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			mp.CreateIndex()
		}
	}

	dbmap = IndependentDbMapper(db)
	dbmap.AddTable(mdl.Community{})
	dbmap.AddTable(mdl.Pdf{})
	dbmap.AddTable(mdl.Page{})
	dbmap.AddTable(mdl.Decision{})

	return dbmap

}

// checkRes is checking the error *and* the sql result
// of a sql query.
func CheckRes(sqlRes sql.Result, err error) {
	defer logx.SL().Incr().Decr()
	defer logx.SL().Incr().Decr()
	util.CheckErr(err)
	liId, err := sqlRes.LastInsertId()
	util.CheckErr(err)
	affected, err := sqlRes.RowsAffected()
	util.CheckErr(err)
	if affected > 0 && liId > 0 {
		logx.Printf("%d row(s) affected ; lastInsertId %d ", affected, liId)
	} else if affected > 0 {
		logx.Printf("%d row(s) affected", affected)
	} else if liId > 0 {
		logx.Printf("%d lastInsertId", liId)
	}
}

func TableName(i interface{}) string {
	t := reflect.TypeOf(i)
	if table, err := dbmap.TableFor(t, false); table != nil && err == nil {
		return dbmap.Dialect.QuoteField(table.TableName)
	}
	return t.Name()
}

func IndependentDbMapper(db *sql.DB) *gorp.DbMap {
	var dbmap *gorp.DbMap
	if config.Config.SQLite {
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
		// We have to enable foreign_keys for EVERY connection
		// There is a gorp pull request, implementing this
		hasFK1, err := dbmap.SelectStr("PRAGMA foreign_keys")
		logx.Printf("PRAGMA foreign_keys is %v | err is %v", hasFK1, err)
		dbmap.Exec("PRAGMA foreign_keys = true")
		hasFK2, err := dbmap.SelectStr("PRAGMA foreign_keys")
		logx.Printf("PRAGMA foreign_keys is %v | err is %v", hasFK2, err)
	} else {
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	}
	return dbmap
}

func TraceOn() {
	DBMap().TraceOn("gorp: ", log.New(os.Stdout, "", 0))
}
func TraceOff() {
	DBMap().TraceOff()
}
