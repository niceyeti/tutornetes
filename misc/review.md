# intro

For debugging and managing cloud applications, containers, service mesh, etc.,
there are some important protocols and other topics to review.

# DNS
DNS provides stable naming conventions for services within a network: name -> IP.
DNS forms a tree/hierarchy, starting with the root zone, and proceeding down to individual IPs and subzones.
The most important aspect of DNS in one's cluster is simply being aware of its implementation: trust boundaries,
what DNS servers are operating, and how clients are configured to use it. In vanilla pod containers, this is
usually in /etc/resolv.conf, though ISTIO surely has a more featured implementation in the Enovy sidecar.

### Record types
1. A/AAAA: the most common type, mapping a name to a specific ip address.
    * Normal k8s Services are listed by these records; 'headless' service records include multiple addresses.
2. PTR: reverse records, using the special domain 'in-addr.arpa', for reverse lookups.
3. SRV: these map to specific named ports/protocols.
4. MX: mail records
5. With regard to k8s services and headless services:
```
Usually, when you perform a DNS lookup for a service, the DNS server returns a single IP — the service’s
cluster IP. But if you tell Kubernetes you don’t need a cluster IP for your service (you do this by setting
the clusterIP field to None in the service specification ), the DNS server will return the pod IPs instead
of the single service IP. Instead of returning a single DNS A record, the DNS server will return multiple
A records for the service, each pointing to the IP of an individual pod backing the service at that moment.
Clients can therefore do a simple DNS A record lookup and get the IPs of all the pods that are part of the
service. The client can then use that information to connect to one, many, or all of them.
```

Address resolution procedure, from right to left of the url:
1. Ask root ("."): "who is www.wikipedia.org?"
2. Root responds: "Try 202.123.123.345" (".org")
3. Ask ".org": "who is www.wikipedia.org?"
4. "Try 123.234.456.678" (nameserver1.wikipedia.org)
5. As nameserver: "who is wikipedia.org?"
6. "wikipedia.org is at 123.123.134.19" (AA record response)
DNS queries are usually cached, instead of being forwarded to root.

Reverse lookup procedure:
Same as forward, but operates via PTR records with the domain name reversed.
For example a forward lookup for 'example.com' may return 192.0.1.2, for which
the reverse lookup PTR name would be '2.1.0.192.in-addr.arpa.'

Security posture:
* Hosts are configured via /etc/resolv.conf
* Hosts' domains are registered, allowing them to be looked up using reverse DNS
* Traceroute uses PTR records to work and display hosts
* Both forward and reverse records are required for many internal network functionalities: spam checking emails,
traceroute, etc.

### nslookup
1. Lookup an ip using builting resolver (see /etc/resolv.conf):
* `nslookup example.com`
2. Lookup an ip, specifying the name server:
* `nslookup example.com resolver1.opendns.com`
* `nslookup example.com [cluster dns server addr]`

### dig
1. Forward lookup: `dig example.com`
2. Reverse lookup: `dig -x 123.234.345.456`

### dnssec
Extends unsecured dns with validation of records by both the server and the client.
Provides integrity, but not confidentiality, since DNS is unencrypted.
PKI: entails key distribution and maintenance for verification of RRSIG entries.
* RRSIG: the signature of a record
* DNSKEY: the public key used to verify the RRSIG


# mTLS




# X509
Binds an identity (hostname, organization, or individual) to a public key using a digital signature from some other authority (unless self-signed),
as well as a duration.
Certs contain: version number, serial number, issuer name, validity period (not before, not after), subject name, subject public key and algorithm, cert signature and algorithm.

### Generation
openssl genrsa -aes128 -out privkey.pem 2048
openssl req -new -x509 -key privkey.pem -out https.cert -days 365 -subj "/C=US/ST=CA/O=Acme, Inc./CN=example.com"






