# GophKeeper

A client-server e2e secrets manager to store:
- login/password
- text
- binary
- bank card info

Secrets are encrypted with Argon2id - AES-256-GCM client-side, the server
only ever stores ciphertext it cannot read without your master password.

**Live demo:** [Swagger UI](https://gophkeeper.vla8islav-personal.cloud/swagger/index.html#/system/get_api_ping) on a deployed
instance. Browse and try the API (register to get a token, then **Authorize**
with `Bearer <token>`).

- **Server** (`cmd/gophkeeper-server`): Go + chi + PostgreSQL (goose migrations).
  Auth (bcrypt + JWT), per-user KDF salt, secret CRUD + sync, optional audit log
- **Client** (`cmd/gophkeeper-client`): stateless CLI with a local SQLite read
  cache (offline reads), commands:
  - `register`
  - `login`
  - `add`
  - `get`
  - `list`
  - `update` 
  - `delete` 
  - `sync`

## Local development

Server (with a local mkcert cert for HTTPS):

```sh
mkcert -install
mkdir -p certs
mkcert -cert-file certs/server.pem -key-file certs/server-key.pem localhost 127.0.0.1 ::1
go run ./cmd/gophkeeper-server -public-key certs/server.pem -private-key certs/server-key.pem
```

Client on the local server (mkcert CA is trusted by the system store):

```sh
go build -o gophkeeper ./cmd/gophkeeper-client
./gophkeeper register alice
./gophkeeper add login_password "mypass"
./gophkeeper list
```

Run the tests(needs Docker):

```sh
go test ./...
```

## API docs (Swagger)

The server serves an interactive Swagger UI (generated from the handler
annotations with [swaggo/swag](https://github.com/swaggo/swag)):

- UI: `http(s)://<host>/swagger/index.html`
- OpenAPI spec: `http(s)://<host>/swagger/doc.json`

Click **Authorize** and paste `Bearer <token>` (the token returned by
`/api/user/register` or `/api/user/login`) to call the authenticated endpoints.

Regenerate the spec after changing any annotation:

```sh
swag init -g cmd/gophkeeper-server/main.go -o docs --parseInternal --parseDependency
```

## Production deployment

The deployment runs three services: 
- **Caddy** that auto-provisions and renews a Let's Encrypt cert for your domain, terminates TLS. 
- **gophkeeper** server (plain HTTP behind Caddy), and **Postgres**.
Assuming you're using a cheap VPS with the external ipv4 and you own the domain, you should:

1. **DNS**: point your domain's A/AAAA record at the VPS. Open ports **80** and
   **443** in the firewall
2. **Config**: `cp .env.example .env` and fill it in:
   1. `DOMAIN`
   2. the Postgres credentials
   3. a strong `AUTH_TOKEN_SECRET` (`openssl rand -hex 32`)
3. **Certs — test first**: to avoid Let's Encrypt rate limits, temporarily enable
   the staging CA in `Caddyfile` (see the commented `acme_ca` line), bring the
   stack up, confirm it works, then remove that line and restart to get a real
   cert.
4. **Launch**:

   ```sh
   docker compose up -d --build
   docker compose logs -f caddy    # watch the cert get issued
   ```

   Migrations run automatically on server startup (goose, in-process).

5. **Use the client** against the deployed server — no `--ca` needed, since the
   Let's Encrypt root is in the system trust store:

   ```sh
   ./gophkeeper -base-url https://your-domain.example.com register alice
   ```

### Audit log

Set `AUDIT_LOG_PATH` (already wired in `docker-compose.yml` to a persisted
volume) to record one JSON-lines event per request — operation, user id, secret
id, client IP(honors `X-Forwarded-For` behind Caddy), status, and timestamp.
Inspect it with:

```sh
docker compose exec gophkeeper cat /var/log/gophkeeper/audit.jsonl
```
