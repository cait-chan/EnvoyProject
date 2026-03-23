# TLS material

- **Downstream (browser → Envoy on :10001):** `server.crt` / `server.key` — generate with your existing `openssl.cnf` for `localhost` (see project docs).
- **Upstream (Envoy → Go on :3001–3003):** run from repo root:

  ```bash
  chmod +x scripts/gen-upstream-tls-certs.sh
  ./scripts/gen-upstream-tls-certs.sh
  ```

  This creates:
  - `upstream-ca.crt` / `upstream-ca.key` — CA (Envoy trusts `upstream-ca.crt` only).
  - `upstream-server.crt` / `upstream-server.key` — cert presented by all three Go services (SAN includes `host.docker.internal`).

Private keys (`*.key`) are gitignored.

## Troubleshooting

- **`WRONG_VERSION_NUMBER` from Envoy** — Almost always means Envoy is speaking **TLS** to an upstream that is still **plain HTTP** (old `go run` still bound to the port). Stop processes on **3001–3003** and restart the three services; confirm logs show **`HTTPS/TLS`** and printed **`cert=`** paths.
- **`curl: ... tlsv1 alert protocol version` (macOS LibreSSL)** — Same cause (HTTPS to a server that isn’t doing TLS), or try **`curl --tlsv1.2 -k https://127.0.0.1:3001/users`**. Check TLS with: **`openssl s_client -connect 127.0.0.1:3001 -servername localhost -tls1_2`** (you should see a certificate chain and `HTTP/1.1` only after typing `GET /health HTTP/1.1` etc.).
