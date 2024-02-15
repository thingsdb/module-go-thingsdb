// ThingsDB module for communication with ThingsDB.
//
// For example:
//
//		// Create the module (@thingsdb scope)
//		new_module('thingsdb', 'github.com/thingsdb/module-go-thingsdb');
//
//		// Configure the module
//		set_module_conf('thingsdb', {
//		    username: 'admin',
//		    password: 'pass',
//	     host: 'localhost',
//	     default_scope: '//stuff',
//		});
//
//		// Use the module
//		thingsdb.query(".id()").then(|res| {
//		    res;  // just return the response.
//		});
package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	thingsdb "github.com/thingsdb/go-thingsdb"
	timod "github.com/thingsdb/go-timod"
	"github.com/vmihailenco/msgpack"
)

var conn *thingsdb.Conn = nil
var mux sync.RWMutex

type nodes struct {
	Host string  `msgpack:"host"`
	Port *uint16 `msgpack:"port"`
}

type config struct {
	Username   *string `msgpack:"username"`
	Password   *string `msgpack:"password"`
	Token      *string `msgpack:"token"`
	Host       string  `msgpack:"host"`
	Port       *uint16 `msgpack:"port"`
	UseSSL     *bool   `msgpack:"use_ssl"`
	SkipVerify *bool   `msgpack:"skip_verify"`
	Timeout    *uint16 `msgpack:"timeout"`
	Nodes      []nodes `msgpack:"nodes"`
}

type request struct {
	Scope string                 `msgpack:"scope"`
	Code  *string                `msgpack:"code"`
	Name  *string                `msgpack:"name"`
	Vars  map[string]interface{} `msgpack:"vars"`
	Args  interface{}            `msgpack:"args"`
}

func handleConf(cfg *config) {
	mux.Lock()
	defer mux.Unlock()

	var port uint16 = 9200
	var tlsConfig *tls.Config = nil

	if cfg.Port == nil {
		port = *cfg.Port
	}

	if cfg.UseSSL != nil && *cfg.UseSSL {
		skipVerify := false
		if cfg.SkipVerify != nil && *cfg.SkipVerify {
			skipVerify = true
		}
		tlsConfig = &tls.Config{
			InsecureSkipVerify: skipVerify,
		}
	}

	if cfg.Username == nil && cfg.Token == nil {
		log.Println("Missing username or token")
		timod.WriteConfErr()
		return
	}

	if cfg.Username != nil && cfg.Token != nil {
		log.Println("Use a username or token, not both")
		timod.WriteConfErr()
		return
	}

	if cfg.Username != nil && cfg.Password == nil {
		log.Println("Missing password for the username")
		timod.WriteConfErr()
		return
	}

	if conn != nil {
		conn.Close()
	}

	conn := thingsdb.NewConn(cfg.Host, port, tlsConfig)
	conn.DefaultTimeout = 10 * time.Second

	if cfg.Timeout != nil {
		conn.DefaultTimeout = time.Duration(*cfg.Timeout) * time.Second
	}

	if cfg.Nodes != nil && len(cfg.Nodes) > 0 {
		for i := 0; i < len(cfg.Nodes); i++ {
			var port uint16 = 9200
			node := cfg.Nodes[i]
			if node.Port == nil {
				port = *node.Port
			}
			conn.AddNode(node.Host, port)
		}
	}

	if err := conn.Connect(); err != nil {
		conn = nil
		log.Println(err)
		timod.WriteConfErr()
		return
	}

	if cfg.Token != nil {
		if err := conn.AuthToken(*cfg.Token); err != nil {
			conn.Close()
			conn = nil
			log.Println(err)
			timod.WriteConfErr()
			return
		}
	} else {
		if err := conn.AuthPassword(*cfg.Username, *cfg.Password); err != nil {
			conn.Close()
			conn = nil
			log.Println(err)
			timod.WriteConfErr()
			return
		}
	}

	timod.WriteConfOk()
}

func onResponse(pkg *timod.Pkg, res []byte, err error) {
	if err != nil {
		if terr, ok := err.(*thingsdb.TiError); ok {
			timod.WriteEx(
				pkg.Pid,
				timod.Ex(terr.Code()),
				terr.Error())
			return
		}
		timod.WriteEx(
			pkg.Pid,
			timod.ExBadData,
			fmt.Sprintf("Unexpected error: %s", err))
		return
	}
	timod.WriteResponseRaw(pkg.Pid, res)
}

func onModuleReq(pkg *timod.Pkg) {
	mux.RLock()
	defer mux.RUnlock()

	if conn == nil {
		timod.WriteEx(
			pkg.Pid,
			timod.ExCancelled,
			"No connection, make sure you used set_module_conf() to initialze the module")
		return
	}

	var req request
	err := msgpack.Unmarshal(pkg.Data, &req)
	if err != nil {
		timod.WriteEx(
			pkg.Pid,
			timod.ExBadData,
			fmt.Sprintf("Invalid request: %s", err))
		return
	}
	if req.Code == nil && req.Name == nil {
		timod.WriteEx(
			pkg.Pid,
			timod.ExBadData,
			"Missing code (for QUERY) or name (for RUN)")
		return
	}

	if req.Code != nil && req.Name != nil {
		timod.WriteEx(
			pkg.Pid,
			timod.ExBadData,
			"Both code (for QUERY) and name (for RUN) are set")
		return
	}

	if req.Code != nil {
		if req.Args != nil {
			timod.WriteEx(
				pkg.Pid,
				timod.ExBadData,
				"Found `args`, use `vars` for QUERY")
			return
		}
		res, err := conn.QueryRaw(req.Scope, *req.Code, req.Vars)
		onResponse(pkg, res, err)
		return
	}

	if req.Name != nil {
		if req.Vars != nil {
			timod.WriteEx(
				pkg.Pid,
				timod.ExBadData,
				"Found `vars`, use `args` for RUN")
			return
		}
		res, err := conn.RunRaw(req.Scope, *req.Name, req.Args)
		onResponse(pkg, res, err)
		return
	}

	timod.WriteEx(
		pkg.Pid,
		timod.ExCancelled,
		"Code should not be reached")
}

func handler(buf *timod.Buffer, quit chan bool) {
	for {
		select {
		case pkg := <-buf.PkgCh:
			switch timod.Proto(pkg.Tp) {
			case timod.ProtoModuleConf:
				var cfg config
				err := msgpack.Unmarshal(pkg.Data, &cfg)
				if err == nil {
					handleConf(&cfg)
				} else {
					log.Println("Error: Missing or invalid ThingsDB configuration")
					timod.WriteConfErr()
				}

			case timod.ProtoModuleReq:
				onModuleReq(pkg)

			default:
				log.Printf("Error: Unexpected package type: %d", pkg.Tp)
			}
		case err := <-buf.ErrCh:
			// In case of an error you probably want to quit the module.
			// ThingsDB will try to restart the module a few times if this
			// happens.
			log.Printf("Error: %s", err)
			quit <- true
		}
	}
}

func main() {
	// Starts the module
	timod.StartModule("thingsdb", handler)

	if conn != nil {
		conn.Close()
	}
}
