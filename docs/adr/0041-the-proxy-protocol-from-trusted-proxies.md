# 41. Cahut reads the PROXY protocol from the proxies it trusts

- Status: Accepted
- Date: 2026-10-25

## Context

[ADR 2](0002-built-in-https-with-certmagic.md) has the binary serve HTTPS itself, on ports 80 and 443, and left a mode behind a proxy for later. That later has come: on the server of my.cahut.cloud, nginx holds port 443 and routes each connection by the name that the client asks for (SNI), without decrypting it. It serves the sign-up site of the host itself, and passes every other connection to Cahut, on a port of the machine. TLS stays between the client and Cahut, and CertMagic gets its certificates as before.

Behind such a proxy, every connection comes from the proxy, at 127.0.0.1. Cahut counts its limits by the address of each client: the calls that sign in, the registrations of OAuth clients, the calls to the token endpoint, the Keys typed. Its logs name the address of each client, and the Logbook that of each visitor. All of them now see one address: the people of the Cahut share 10 sign-ins, then one every 6 seconds, and the logs say 127.0.0.1. A host cannot open a Cahut so to the directories of AI clients.

A proxy that does not decrypt cannot add a header to the HTTP requests it carries. The PROXY protocol, written for HAProxy and spoken by nginx, Traefik, Envoy and the load balancers of cloud providers, puts the address of the client in a header before the first byte of the connection: a line of text in version 1, a binary block in version 2.

Whoever sends that header chooses the address that Cahut sees. It must be read from the proxies only, and from them always.

## Decision

### The settings

- `cahut serve --https-proxy-protocol <addresses>`, or `CAHUT_HTTPS_PROXY_PROTOCOL`, names the proxies whose PROXY protocol Cahut reads on its HTTPS listener: IP addresses or prefixes, separated by commas, such as `127.0.0.1` or `10.0.0.0/8`. `--http-proxy-protocol`, or `CAHUT_HTTP_PROXY_PROTOCOL`, does the same on its HTTP listener. Both are off when empty, the default. Cahut refuses to start with an entry that is neither an address nor a prefix.
- Each listener has its own list, since a proxy can route one port in TCP and the other in HTTP. nginx on the server of my.cahut.cloud routes port 443 by SNI, but port 80 by the `Host` header, and its HTTP proxy cannot send the PROXY protocol.
- Neither setting applies to `--dev`, nor to the admin API, which serves the machine itself by the address of each connection ([ADR 35](0035-a-program-on-the-server-manages-people-with-a-service-token.md)).

### Each connection

- From a trusted proxy, a connection must start with a header of the PROXY protocol, version 1 or 2, within 10 seconds, before TLS or HTTP. Cahut takes the address and the port of the client from it, and the address that the client called as the local address of the connection. Cahut then serves the connection as if the client had called it directly.
- A header of version 1 with `UNKNOWN`, or of version 2 with the `LOCAL` command or with an address that is neither IPv4 nor IPv6, is the proxy speaking for itself, as its health checks do: the connection keeps the address of the proxy. Cahut reads the extensions of version 2 (its TLVs) and sets them aside.
- Without a valid header, Cahut closes the connection, serves nothing on it, and logs why, with the address of the proxy. No connection of a trusted proxy falls back to the address of the proxy. A connection that closes before its first byte, as the TCP checks of load balancers do, closes without a line.
- From any other address, Cahut reads no header: the connection is served as it comes, under its own address. A header that such a client sends is the start of a TLS handshake or of an HTTP request, which fails.
- So a trusted address must reach its listener through the proxy only. Behind nginx on the same machine, Cahut listens on 127.0.0.1, where only the programs of the machine connect.

### What the address of the client changes

- The limits, the logs and the Logbook take the address that the proxy passes, as they took that of the connection: a client of IPv6 still counts by its /64.
- The header comes before TLS, so the TLS-ALPN-01 challenge of Let's Encrypt, which comes through the proxy like any other connection, keeps working, as the HTTP-01 challenge on port 80 does. A Cahut at the IP address of its server ([ADR 26](0026-cahut-at-an-ip-address.md)), whose clients send no server name, gets its certificate from the address that the client called, which the header names.
- The local health check ([ADR 24](0024-signed-releases-and-updates.md)) answers on the HTTP listener to a client at a loopback address, for a loopback host. Through a proxy that passes the PROXY protocol, that is the address of the client, so a client of the internet never gets it. When Cahut holds port 80 itself, the updater and `cahut doctor` connect to it directly, without a header: 127.0.0.1 must not be trusted on HTTP then, or the health check of each update fails, and the update rolls back.
- The local health check also says, for each listener that reads the PROXY protocol, the proxies it trusts, the last connection of a proxy that it took and the last one it refused, with why. `cahut doctor` says when the last one was refused: the proxy does not send the header, or not a valid one.

## Consequences

- A Cahut behind nginx `stream`, HAProxy or the load balancer of a cloud provider counts its limits and writes its logs by the address of each client, with TLS and certificates still its own. On the server of my.cahut.cloud, the second server of nginx passes the PROXY protocol on instead of stripping it, and Cahut trusts 127.0.0.1 on HTTPS.
- The proxy and Cahut change together. A proxy that sends the header to a Cahut that does not read it breaks every TLS handshake, and a Cahut that waits for the header refuses every connection of a proxy that does not send it. `cahut doctor` says which side is missing.
- On port 80, the clients that come through an HTTP proxy keep the address of the proxy. Port 80 answers the challenges of Let's Encrypt and redirects to HTTPS: it counts no client, and logs none. But every client of such a proxy is then at 127.0.0.1, and the local health check would answer any of them that names a loopback host: the proxy lets `/-/healthz` through from the machine only, as the documentation of the service shows.
- `install.sh` keeps both settings when it writes `/etc/cahut/cahut.env` again. The addresses that Cahut listens on behind a proxy stay in an override of the systemd unit, as before.
- The code reads both versions of the protocol with the standard library, in one file of `internal/web`.

## Alternatives considered

- **`X-Forwarded-For` or `Forwarded` from trusted proxies.** Only a proxy that decrypts TLS can add them: the certificates and the keys would leave Cahut for the proxy.
- **A header optional from trusted proxies.** A proxy that passes a connection without a header would let its client write one, and choose its own address.
- **One list for both listeners.** The HTTP proxy of nginx cannot send the header, so its connections to port 80 would all be refused, unless the header were optional.
- **The library `github.com/pires/go-proxyproto`**, which Traefik and Caddy use. A dependency, with policies that Cahut does not need, for a parser that the specification fixes and that the tests cover.
- **Higher limits.** They would hold back nobody, and the logs would still say 127.0.0.1.
