# Generating a TLS certificate.

Create a directory at the root of the project to store the keys:

```sh
mkdir tls & cd tls
```

Find out where Go is installed:

```sh
which go
```

It should yield something like this, take note of this path:

```
/usr/local/go-1.23/bin/go
```

Use it to call `generate_cert.go`:

```sh
go run /usr/local/go-1.23/bin/go/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost
```

If on mac it's probably something like:

```sh
go run /opt/homebrew/Cellar/go/<version>/libexec/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost
```

It should generate a pair of `cert.pem` and `key.pem` files.
