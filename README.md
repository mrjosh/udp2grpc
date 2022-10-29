![License](https://img.shields.io/github/license/mrjosh/udp2grpc)

<pre align="center">
   __  ______  ____ ___         ____  ____  ______
  / / / / __ \/ __ \__ \ ____  / __ \/ __ \/ ____/
 / / / / / / / /_/ /_/ // __ \/ /_/ / /_/ / /     
/ /_/ / /_/ / ____/ __// /_/ / _, _/ ____/ /___   
\____/_____/_/   /____/\__, /_/ |_/_/    \____/   
/____/
</pre>

## ⚠️ This project is still in early development. ⚠️

## UDP2gRPC
A Tunnel which Turns UDP Traffic into Encrypted gRPC/TCP Traffic,helps you Bypass UDP FireWalls(or Unstable UDP Environment)
Assume your UDP is blocked or being QOS-ed or just poorly supported.

## Getting Started
### Installing
Download binary release from https://github.com/mrjosh/udp2grpc/releases

### Generate certificates for server and client
Assume your server ip is 127.0.0.1 and your service domain is example.com
```bash
# generate for specific ip address
utg gen-certificates --dir ./cert --ip 127.0.0.1

# generate for specific domain name
utg gen-certificates --dir ./cert --domain example.com

# generate for both domain and ip
utg gen-certificates --dir ./cert --domain example.com --ip 127.0.0.1
```

### Running
Assume your server domain example.com and you have a service listening on udp port 51820.
if you wish to run the server without tls, use the flag `--insecure` for client and server
```bash
# Run at server side:
utg server -l0.0.0.0:52935 -r127.0.0.1:51820 --password="super-secure-password" --tls-cert-file cert/server.crt --tls-key-file cert/server.key

# Run at client side:
utg client -rexample.com:52935 -l0.0.0.0:51820 --password="super-secure-password" --tls-cert-file cert/server.crt 
```

### Docker-Compose example
```yaml
version: '3.7'

services:

  # init-container
  # generate certifiactes for server and client
  gen-certificates:
    image: mrjoshlab/udp2grpc:latest
    command:
      - "gen-certificates"
      # server ip address
      - "--ip"
      - "127.0.0.1"
      # certificates directory
      - "--dir"
      - "/cert"
    volumes:
      - "$PWD/cert/:/cert"

  # udp2grpc server container
  udp2grpc-server:
    image: mrjoshlab/udp2grpc:latest
    ports:
      - "52935:52935/tcp"
    command:
      - "server"
      # grpc listen address
      - "-l0.0.0.0:52935"
      # remote conn address
      - "-r127.0.0.1:51820"
      # tls certificate public file
      - "--tls-cert-file"
      - "/cert/server.crt"
      # tls certificate pivate file
      - "--tls-key-file"
      - "/cert/server.key"
      # super secure password here
      - "--password=super-secure-password"
    volumes:
      - "$PWD/cert/:/cert"
    restart: unless-stopped
    depends_on:
      gen-certificates:
        condition: service_completed_successfully

  # udp2grpc client container
  udp2grpc-client:
    image: mrjoshlab/udp2grpc:latest
    ports:
      - "51820:51820/udp"
    command:
      - "client"
      # local udp connection address
      - "-l0.0.0.0:51820"
      # server ip address with port
      - "-r127.0.0.1:52935"
      # tls certificate public file
      - "--tls-cert-file"
      - "/cert/server.crt"
      # super secure password here
      - "--password=super-secure-password"
    volumes:
      - "$PWD/cert/server.crt:/cert/server.crt"
    restart: unless-stopped
    depends_on:
      gen-certificates:
        condition: service_completed_successfully
```

## Contributing
Thank you for considering contributing to UDP2gRPC project!

## License
The UDP2gRPC is open-source software licensed under the MIT license.
