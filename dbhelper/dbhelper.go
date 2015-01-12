// dbhelper.go
package dbhelper

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
)

type Processlist struct {
	Id       uint64
	User     string
	Host     string
	Database sql.NullString
	Command  string
	Time     float64
	State    string
}

type SlaveHosts struct {
	Server_id uint64
	Host      string
	Port      uint
	Master_id uint64
}

type SlaveStatus struct {
	Slave_IO_State                string
	Master_Host                   string
	Master_User                   string
	Master_Port                   uint
	Connect_Retry                 uint
	Master_Log_File               string
	Read_Master_Log_Pos           uint
	Relay_Log_File                string
	Relay_Log_Pos                 uint
	Relay_Master_Log_File         string
	Slave_IO_Running              string
	Slave_SQL_Running             string
	Replicate_Do_DB               string
	Replicate_Ignore_DB           string
	Replicate_Do_Table            string
	Replicate_Ignore_Table        string
	Replicate_Wild_Do_Table       string
	Replicate_Wild_Ignore_Table   string
	Last_Errno                    uint
	Last_Error                    string
	Skip_Counter                  uint
	Exec_Master_Log_Pos           uint
	Relay_Log_Space               uint
	Until_Condition               string
	Until_Log_File                string
	Until_Log_Pos                 uint
	Master_SSL_Allowed            string
	Master_SSL_CA_File            string
	Master_SSL_CA_Path            string
	Master_SSL_Cert               string
	Master_SSL_Cipher             string
	Master_SSL_Key                string
	Seconds_Behind_Master         sql.NullInt64
	Master_SSL_Verify_Server_Cert string
	Last_IO_Errno                 uint
	Last_IO_Error                 string
	Last_SQL_Errno                uint
	Last_SQL_Error                string
	Replicate_Ignore_Server_Ids   string
	Master_Server_Id              uint
	Master_SSL_Crl                string
	Master_SSL_Crlpath            string
	Using_Gtid                    string
	Gtid_IO_Pos                   string
}

func Connect(user string, password string, address string) *sqlx.DB {
	db, _ := sqlx.Open("mysql", user+":"+password+"@"+address+"/")
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func GetAddress(host string, port string, socket string) string {
	var address string
	if host != "" {
		address = "tcp(" + host + ":" + port + ")"
	} else {
		address = "unix(" + socket + ")"
	}
	return address
}

func GetProcesslist(db *sqlx.DB) []Processlist {
	pl := []Processlist{}
	err := db.Select(&pl, "SELECT id, user, host, `db` AS `database`, command, time_ms as time, state FROM INFORMATION_SCHEMA.PROCESSLIST")
	if err != nil {
		log.Fatal(err)
	}
	return pl
}

func GetSlaveStatus(db *sqlx.DB) SlaveStatus {
	db.MapperFunc(strings.Title)
	ss := SlaveStatus{}
	db.Get(&ss, "SHOW SLAVE STATUS")
	/* TBD: Improve handling of empty show status
	 if err != nil {
		log.Fatal(err)
	} */
	return ss
}

func GetSlaveHosts(db *sqlx.DB) map[string]interface{} {
	rows, err := db.Queryx("SHOW SLAVE HOSTS")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	results := make(map[string]interface{})
	for rows.Next() {
		err = rows.MapScan(results)
		if err != nil {
			log.Fatal(err)
		}
	}
	return results
}

func GetSlaveHostsArray(db *sqlx.DB) []SlaveHosts {
	sh := []SlaveHosts{}
	err := db.Select(&sh, "SHOW SLAVE HOSTS")
	if err != nil {
		log.Fatal(err)
	}
	return sh
}

func GetSlaveHostsDiscovery(db *sqlx.DB) []string {
	slaveList := []string{}
	/* This method does not return the server ports, so we cannot rely on it for the time being. */
	err := db.Select(&slaveList, "select host from information_schema.processlist where command ='binlog dump'")
	if err != nil {
		log.Fatal(err)
	}
	return slaveList
}

func GetStatus(db *sqlx.DB) map[string]string {
	type Variable struct {
		Variable_name string
		Value         string
	}
	vars := make(map[string]string)
	rows, err := db.Queryx("SELECT Variable_name AS variable_name, Variable_Value AS value FROM information_schema.global_status")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var v Variable
		err := rows.Scan(&v.Variable_name, &v.Value)
		if err != nil {
			log.Fatal(err)
		}
		vars[v.Variable_name] = v.Value
	}
	return vars
}

func GetStatusAsInt(db *sqlx.DB) map[string]int64 {
	type Variable struct {
		Variable_name string
		Value         int64
	}
	vars := make(map[string]int64)
	rows, err := db.Queryx("SELECT Variable_name AS variable_name, Variable_Value AS value FROM information_schema.global_status")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var v Variable
		rows.Scan(&v.Variable_name, &v.Value)
		vars[v.Variable_name] = v.Value
	}
	return vars
}

func GetVariables(db *sqlx.DB) map[string]string {
	type Variable struct {
		Variable_name string
		Value         string
	}
	vars := make(map[string]string)
	rows, err := db.Queryx("SELECT Variable_name AS variable_name, Variable_Value AS value FROM information_schema.global_variables")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var v Variable
		err := rows.Scan(&v.Variable_name, &v.Value)
		if err != nil {
			log.Fatal(err)
		}
		vars[v.Variable_name] = v.Value
	}
	return vars
}

func GetVariableByName(db *sqlx.DB, name string) string {
	var value string
	err := db.QueryRowx("SELECT Variable_Value AS Value FROM information_schema.global_variables WHERE Variable_Name = ?", name).Scan(&value)
	if err != nil {
		log.Fatal(err)
	}
	return value
}