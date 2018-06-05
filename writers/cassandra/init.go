package cassandra

import "github.com/gocql/gocql"

const table = `CREATE TABLE IF NOT EXISTS messages (
        id uuid PRIMARY KEY,
        channel bigint,
    	publisher bigint,
        protocol text,
    	name text,
    	unit text,
    	value double,
    	string_value text,
        bool_value boolean,
        data_value text,
    	value_sum double,
    	time double,
    	update_time double,
    	link text,
    )`

// Connect establishes connection to the Cassandra cluster.
func Connect(hosts []string, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum

	return cluster.CreateSession()
}

// Initialize creates table used by the writer service.
func Initialize(session *gocql.Session) error {
	return session.Query(table).Exec()
}
