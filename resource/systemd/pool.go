// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package systemd

import "github.com/coreos/go-systemd/dbus"

// Conn wrapps a dbus connection
type Conn struct {
	Connection *dbus.Conn
}

// Return returns the dbus connection to the connection pool. If the pool is full,
// close the connection.
func (c *Conn) Return() {
	if c.Connection != nil {
		select {
		case connectionPool <- c.Connection:
			c.Connection = nil
		default:
			c.Connection.Close()
			c.Connection = nil
		}
	}
}

var connectionPool chan *dbus.Conn

// ConnectionPoolSize defines how many connections should be in the pool at
// any one time
const ConnectionPoolSize int = 5

func init() {
	connectionPool = make(chan *dbus.Conn, ConnectionPoolSize)
}

// GetDbusConnection gets a systemd.Conn. This wrapps the dbusConn. It is necessary to call `.Return`
// when you are done with the connection.
func GetDbusConnection() (*Conn, error) {
	for {
		select {
		case dbusConn := <-connectionPool:
			conn := &Conn{Connection: dbusConn}
			return conn, nil
		default:
			dbusConn, err := dbus.New()
			if err != nil {
				return nil, err
			}
			select {
			case connectionPool <- dbusConn:
			default:
				dbusConn.Close()
			}
		}
	}
}
