package main


import (
    "time"
    "os"
    "log"
    "fmt"
    "net"
    "golang.org/x/net/icmp"
    "golang.org/x/net/ipv4"
	"strconv"
)

const (

    ProtocolICMP = 1
    
)


var ListenAddr = "0.0.0.0"
var errcount = 0

func Ping(addr string, ttl int) (*net.IPAddr, time.Duration, error) {

    c, err := icmp.ListenPacket("ip4:icmp", ListenAddr)
    if err != nil {
        return nil, 0, err
    }
    defer c.Close()
	
	
	cd := c.IPv4PacketConn().SetTTL(ttl);
	fmt.Printf("TTL: %d ",ttl)
	if cd!=nil {
		return nil,0,cd
	}

    
	dst, err := net.ResolveIPAddr("ip4", addr)
    if err != nil {
        panic(err)
        return nil, 0, err
    }

 
    m := icmp.Message{
        Type: ipv4.ICMPTypeEcho, Code: 0,
        Body: &icmp.Echo{
            ID: os.Getpid() & 0xffff, Seq: 1, //<< uint(seq), // TODO
            Data: []byte(""),
        },
    }
    b, err := m.Marshal(nil)
    if err != nil {
        return dst, 0, err
    }

    start := time.Now()
    n, err := c.WriteTo(b, dst)
    if err != nil {
        return dst, 0, err
    } else if n != len(b) {
        return dst, 0, fmt.Errorf("got %v; want %v", n, len(b))
    }

    reply := make([]byte, 1500)
    err = c.SetReadDeadline(time.Now().Add(10 * time.Second))
    if err != nil {
        return dst, 0, err
    }
    n, peer, err := c.ReadFrom(reply)
    if err != nil {
        return dst, 0, err
    }
    duration := time.Since(start)


    rm, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
    if err != nil {
        return dst, 0, err
    }
    switch rm.Type {
    case ipv4.ICMPTypeEchoReply:
        return dst, duration, nil
    default:
        return dst, 0, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
    }
}

func main() {
    var total = 0
	var pacloss = 0
	p := func(addr string, ttl int){
        dst, dur, err := Ping(addr , ttl)
        if err != nil {
            errcount = errcount + 1
			log.Printf("Ping %s (%s): %s\n", addr, dst, err)
            return
        }
        log.Printf("Ping %s (%s) RTT: %s\n", addr, dst, dur)
    }
    argsWithProg := os.Args[1]
	ttl := os.Args[2]
	i, err := strconv.Atoi(ttl)
	if err != nil {
            log.Fatal(err)
        }
	
	for{
	time.Sleep(1 * time.Second)
	total = total + 1
	p(argsWithProg, i)
	pacloss = (errcount/total) * 100
	fmt.Printf("Packetloss: %d percent ",pacloss)
	}


}