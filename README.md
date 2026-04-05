# goRacer

<br>

<b>goRacer</b> is a CLI tool for exploiting remote race condition flaws by completing multiple HTTP/2 requests within a single TCP packet.  
For more info, see <a href="https://portswigger.net/research/smashing-the-state-machine#single-packet-attack">Smashing the state machine</a>.


<br>

# Installation Instructions

**Requirements**
 - Go 1.22+

**Installation**

```sh
go install github.com/ri2q/goRacer@latest
```

<br>

# Flags:

```console

BASIC:
   -r, --request	Path to HTTP request file(s)
   -u, --url	Target URL (https://example.com, example.com, example.com:8080)
   -i, --iteration	Concurrent racers per request (default: 25)
   -m, --method	HTTP method  (default: GET)
   -d, --data	Form-encoded body 
   -j, --json	JSON request body
   -x, --header	HTTP headers (key=value)
   -s, --sensitive-header	Secure headers (key=value)
   -c, --cookie	Cookies (key=value)
   -f, --filter	Filter responses by HTTP status code (default: 200)
   -e, --end-stream	Send requests without using signle-packet feature (e.i.,  normal request)

DEBUG:   
   -H, --delay	Delay between iterations in ms (default: 100)
   -l, --tls-key-log	TLS key log file path for wireshark
   -p, --proxy	Proxy (http://host:port or host:port)
```

<br>

# Example Usages:

**Url as input**

```sh
goRacer -u example.com/end-point -i 30
```

 - If you use -d or -j without specifying -m, the default method will be POST.

```sh
goRacer -u https://example.com/end-point -d "key=value"

goRacer -u https://example.com/end-point \
   -c "session=value" \
   -j '{"key":"value"}' \
    -x "Host=example.com"
```

**File as input**

```sh
goRacer -r request.txt -i 30 
```
 - You can use multiple files at the same time:

```sh
goRacer -r r1.txt r2.txt r3.txt -i 30 -f 302
```

**Advanced Configuration**

 - You can fine-tune the delay between the first and last frames (END_STREAM) using -H (in milliseconds):

```sh
goRacer -u example.com/end-point -H 150
```
 - The -e flag disables the last-segment feature. It cannot be used together with -H:

```sh
goRacer -u example.com/end-point -e -i 10  
```

**Debugging**
 - Use an HTTP or SOCKS proxy to inspect requests:
 
```sh
goRacer -r req.txt -p 127.0.0.1:8080   
```
 - Use the -l flag to specify a TLS key log file for Wireshark:

```sh
goRacer -u example.com/end-point -l ./path/to/tlslog.txt
```
