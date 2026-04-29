projects
    - id
        - project name (например -> proxy)
         - .crewai
         - folders/files





Как должно работать создаётся проект те папка сначала идёт её id в ней лежит ещё одна но уже с самим названием потом уже в ней лежит .crewai и тд и внедряеться гит git init после чего каждый воркер либо мэнэджер лусше наверное мэнэджер создаёт ветку каждому воркеру и он в неё работает потом когда они закончили мэнэджер обьеденяет и тестриует да ксати есть баг в файл записались команды который должны были выполниться:
```go 
package main

import (
	"flag"
	"log"
	"os"
	"time"
	
	"github.com/yourusername/yourproject/proxy"
)

func main() {
	// Parse command-line flags
	config := parseFlags()
	
	// Create proxy server
	server, err := proxy.NewProxyServer(config)
	if err != nil {
		log.Fatalf("Failed to create proxy server: %v", err)
	}
	
	// Start server
	if config.TLSEnabled {
		err = server.StartWithTLS()
	} else {
		err = server.Start()
	}
	
	if err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}
}

func parseFlags() *proxy.Config {
	config := proxy.DefaultConfig()
	
	// Basic options
	host := flag.String("host", config.Host, "Host to bind to")
	port := flag.Int("port", config.Port, "Port to listen on")
	
	// Timeouts
	readTimeout := flag.Duration("read-timeout", config.ReadTimeout, "Read timeout")
	writeTimeout := flag.Duration("write-timeout", config.WriteTimeout, "Write timeout")
	idleTimeout := flag.Duration("idle-timeout", config.IdleTimeout, "Idle timeout")
	upstreamTimeout := flag.Duration("upstream-timeout", config.UpstreamTimeout, "Upstream timeout")
	
	// Connection limits
	maxConns := flag.Int("max-connections", config.MaxConnections, "Maximum concurrent connections")
	
	// Upstream configuration
	upstreamHost := flag.String("upstream-host", "", "Upstream proxy host")
	upstreamPort := flag.Int("upstream-port", 0, "Upstream proxy port")
	
	// TLS options
	tlsEnabled := flag.Bool("tls", false, "Enable TLS")
	tlsCert := flag.String("tls-cert", "", "TLS certificate file")
	tlsKey := flag.String("tls-key", "", "TLS key file")
	
	// Help flag
	help := flag.Bool("help", false, "Show help")
	
	flag.Parse()
	
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	
	// Apply configuration
	config.Host = *host
	config.Port = *port
	config.ReadTimeout = *readTimeout
	config.WriteTimeout = *writeTimeout
	config.IdleTimeout = *idleTimeout
	config.UpstreamTimeout = *upstreamTimeout
	config.MaxConnections = *maxConns
	config.UpstreamHost = *upstreamHost
	config.UpstreamPort = *upstreamPort
	config.TLSEnabled = *tlsEnabled
	config.TLSCertFile = *tlsCert
	config.TLSKeyFile = *tlsKey
	
	return config
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
=== COMMANDS ===
mkdir -p proxy
touch proxy/go.mod
echo 'module github.com/yourusername/yourproject

go 1.21' > proxy/go.mod
go mod init github.com/yourusername/yourproject
go mod tidy
go build -o proxy-server .
./proxy-server --port 8080 --host 0.0.0.0
./proxy-server --port 8443 --tls --tls-cert cert.pem --tls-key key.pem
./proxy-server --help
```

--- Но вот за что я похволю то как парситься код все отсупы коментарии синтаксис отлично но с командами надо поработать 


Так вот также воркеры делают коммиты в своих ветках 