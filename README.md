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

## Getting Started
### Installing
Download binary release from https://github.com/mrjosh/udp2grpc/releases

### Running
Assume your UDP is blocked or being QOS-ed or just poorly supported.
Assume your server ip is 44.55.66.77, you have a service listening on udp port 51820.

```bash
# Run at server side:
utg server -l0.0.0.0:52935 -r127.0.0.1:51820 --password="super-secure-password" --tls-cert-file cert/server.crt --tls-key-file cert/server.key

# Run at client side:
utg client -rdomain.tld:52935 -l0.0.0.0:51820 --password="super-secure-password" --tls-cert-file cert/server.crt 
```

if you wish to run the server without tls, use the flag `--insecure` for client and server

## Contributing
Thank you for considering contributing to UDP2gRPC project!

## License
The UDP2gRPC is open-source software licensed under the MIT license.
