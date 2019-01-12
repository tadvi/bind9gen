package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "text/template"
)

var confContent = `
zone "{{ .Domain }}" {
type master;
file "/etc/bind/zones/{{ .Domain }}.db";
};
`

var zonesContent = `$TTL {{ .TTL }}
@ 86400 IN SOA ns1.{{ .Domain }}. admin@{{ .Domain }}. (
    2008021501 ; serial, todays date+todays
    86400 ; refresh, seconds
    7200 ; retry, seconds
    3600000 ; expire, seconds
    86400 ) ; minimum, seconds
{{ .Domain }}. 86400 IN NS ns1.{{ .Domain }}.
{{ .Domain }}. 86400 IN NS ns2.{{ .Domain }}.
ns1 IN A {{ .IP }}
ns2 IN A {{ .IP }}
{{ .Domain }}. IN A {{ .IP }}
localhost.{{ .Domain }}. IN A 127.0.0.1
{{ .Domain }}. IN MX 10 {{ .Domain }}.
mail IN CNAME {{ .Domain }}.
www IN CNAME {{ .Domain }}.
ftp IN A {{ .IP }}
`

const namedFile = "/etc/bind/named.conf.local"
const zonesDir = "/etc/bind/zones"

type Config struct {
    Domain string
    IP     string
    TTL    int
}

var config Config

func init() {
    flag.StringVar(&config.Domain, "domain", "", "domain name")
    flag.StringVar(&config.IP, "ip", "", "IP for the domain")
    flag.IntVar(&config.TTL, "ttl", 3600, "TTL for the domain")
}

func main() {
    flag.Parse()

    fmt.Println("Example of use: ./bind9gen -domain=mydomain.com -ip=192.168.1.1 -ttl=3600")
    if config.Domain == "" || config.IP == "" {
        log.Fatal("Missing input parameters for Domain and IP")
    }

    // Creating named.conf file.
    f, err := os.Create(namedFile)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    tpl, err := template.New("").Parse(confContent)
    if err != nil {
        log.Fatal(err)
    }
    err = tpl.Execute(f, &config)
    if err != nil {
        log.Fatal(err)
    }

    // Creating zones file.
    err = os.MkdirAll(zonesDir, 0777)
    if err != nil {
        log.Fatal(err)
    }

    zf, err := os.Create(filepath.Join(zonesDir, config.Domain+".db"))
    if err != nil {
        log.Fatal(err)
    }
    if err = zf.Chmod(0666); err != nil {
        log.Fatal(err)
    }
    defer zf.Close()

    tpl, err = template.New("").Parse(zonesContent)
    if err != nil {
        log.Fatal(err)
    }
    err = tpl.Execute(zf, &config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Done.")
    fmt.Println("Do not forget to run: systemctl enable bind9")
}
