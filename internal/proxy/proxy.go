package proxy

import (
	"log"
	"net"
	"os"

	"github.com/lyuangg/umyproxy/protocol"
)

type (
    Proxy struct {
        server net.Listener
        pool *Pool
        socketFile string
        debug bool
    }
)

func NewProxy(p *Pool, socketfile string) *Proxy {
    return &Proxy{
        pool: p,
        socketFile: socketfile,
    }
}

func (p *Proxy) Run() {
    p.deleteSocketFile()

    serv, err := net.Listen("unix", p.socketFile)
    if err != nil {
        log.Fatalln("Listen err:", err)
    }
    p.server = serv
    p.startPrint()

    for {
        conn, err := p.server.Accept()
        if err != nil {
            log.Fatalln("conn err:", err)
        }
        p.debugPrintf("accept conn")

        go p.HandleConn(conn)
    }
}

func (p *Proxy) SetDebug() {
    p.debug = true
    p.debugPrintf("debug mode")
}

func (p *Proxy) debugPrintf(format string, v ...interface{}) {
    if p.debug {
        format = "[DEBUG]" + format + "\n"
        log.Printf(format, v...)
    }
}

func (p *Proxy) startPrint() {
    log.Println("start server: ", p.socketFile)
    log.Println("host:", p.pool.option.Host)
    log.Println("port:", p.pool.option.Port)
    log.Println("pool_size:", p.pool.option.PoolMaxSize)
    log.Println("conn_maxlifetime:", p.pool.option.MaxLifetime)
    log.Println("wait_timeout:", p.pool.option.WaitTimeout)
}

func (p *Proxy) HandleConn(conn net.Conn) {
    client := protocol.NewConn(conn)
    defer client.Close()

    mysqlServ, err := p.Get()
    if err != nil {
        log.Printf("get mysql conn err: %+v \n", err)
        return
    }
    p.debugPrintf("get mysql conn")
    defer p.Put(mysqlServ)

    // 认证
    if err := mysqlServ.Auth(client); err != nil {
        log.Printf("mysql auth err: %+v \n", err)
        return
    }
    p.debugPrintf("client auth success")


    // 发送命令
    for {
        cmd, err := client.ReadPacket()

        if err != nil {
            log.Println("read cmd err: ", err)
            return
        }

        p.debugPrintf("read cmd: %+v", cmd)

        if protocol.IsQuitPacket(cmd) {
            p.debugPrintf("client quit")
            return
        }

        err = mysqlServ.WritePacket(cmd)
        if err != nil {
            log.Printf("write cmd to server err: %+v \n", err)
            return
        }

        // response
        resp := protocol.NewResponse(mysqlServ, cmd.Payload[0])
        err = resp.ResponsePacket(client)
        p.debugPrintf("transport response")
        if err != nil {
            log.Println("transport response err:", err)
            return
        }
        p.debugPrintf("end transport response")
    }

}

func (p *Proxy) Get() (protocol.Connector, error) {
    return p.pool.Get()
}

func (p *Proxy) Put(conn protocol.Connector) error {
    p.debugPrintf("put conn")
    return p.pool.Put(conn)
}

func (p *Proxy) Close() {
    p.pool.Close()
    p.server.Close()
}

func (p *Proxy) deleteSocketFile() error {
    _, err := os.Stat(p.socketFile)
    if err == nil || os.IsExist(err) {
        return os.Remove(p.socketFile)
    }
    return err
}
