![code coverage](https://github.com/sucamchi/learn-cicd-starter/actions/workflows/ci.yml/badge.svg)

# Notely

A small note-taking app I built while working through the
[Learn CI/CD](https://boot.dev) course on Boot.dev. It started as the course
starter code and became my own working pipeline: tests, linting, security
scanning, database migrations, a Docker image, and a deploy to Google Cloud Run,
all running from GitHub Actions.

## What the app does

- Serves a single page at `/` where you create a user and add notes. There is no
  password. The API key is kept in the browser's `localStorage`.
- Exposes a small JSON API under `/v1` for users and notes.
- Stores users and notes in [Turso](https://turso.tech), a hosted libSQL
  database, through `sqlc`-generated Go queries.
- Runs without a database if `DATABASE_URL` is missing. In that mode the page
  still loads but the CRUD endpoints are disabled.

## What I learned

**Continuous integration (pull requests)**

- Running `go test ./... -cover` on every pull request.
- Enforcing `gofmt` so formatting never shows up in review.
- Catching bugs with `staticcheck`.
- Scanning for security problems with `gosec`.

**Continuous deployment (pushes to `main`)**

- Building a Linux binary and installing `goose` in the runner.
- Running database migrations from the pipeline, using a GitHub secret for the
  connection string.
- Building and pushing a Docker image to Artifact Registry with Cloud Build.
- Deploying that image to Cloud Run, only after the build and migration pass.
- Authenticating to Google Cloud with Workload Identity Federation, so there are
  no long-lived GCP keys in GitHub.
- Giving the running app its database URL from Secret Manager instead of baking
  it into the image.

**Things that bit me**

- A secret wrapped in quotes fails the migration, because the quotes become part
  of the connection string.
- The Cloud Run runtime service account needs `Secret Manager Secret Accessor`,
  otherwise the revision cannot start.
- Ordering matters: migrate after the image builds, before the deploy, so a
  broken build never touches the schema.

## Stack

| Piece | Used for |
| --- | --- |
| Go 1.22+ | Application |
| chi | HTTP routing |
| Turso / libSQL | Hosted database |
| goose | SQL migrations |
| sqlc | Type-safe queries from SQL |
| Docker | Container image |
| GitHub Actions | CI and CD |
| Artifact Registry | Image storage |
| Cloud Run | Hosting |
| Secret Manager | Database URL at runtime |
| Workload Identity Federation | GitHub to GCP auth |

## Run it locally

You need Go 1.22 or newer and `goose` v3:

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.27.0
```

Create a `.env` file in the project root. It is gitignored, so it should never
show up in a commit:

```bash
PORT=8080
DATABASE_URL=libsql://your-db-yourname.turso.io?authToken=YOUR_TOKEN
```

Get the token with the Turso CLI:

```bash
turso db tokens create your-db
```

Apply the migrations:

```bash
./scripts/migrateup.sh
```

Build and start the server:

```bash
go build -o notely && ./notely
```

Open <http://localhost:8080>. If you leave `DATABASE_URL` out, the app logs
`Running without CRUD endpoints` and only the page and `/v1/healthz` work.

## API

| Method | Path | Auth | Body |
| --- | --- | --- | --- |
| `POST` | `/v1/users` | none | `{"name": "Susana"}` |
| `GET` | `/v1/users` | `ApiKey <key>` | none |
| `POST` | `/v1/notes` | `ApiKey <key>` | `{"note": "Hello"}` |
| `GET` | `/v1/notes` | `ApiKey <key>` | none |
| `GET` | `/v1/healthz` | none | none |

`POST /v1/users` returns the generated API key. Send it back in the
`Authorization` header on the other calls:

```bash
curl -X POST localhost:8080/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"name": "Susana"}'
```

## How the pipelines work

`.github/workflows/ci.yml` runs on every pull request. The `tests` job runs the
test suite, then `gosec`. The `style` job checks `gofmt` and runs
`staticcheck`.

`.github/workflows/cd.yml` runs on every push to `main`:

1. Check out the code, set up Go, install `goose`.
2. Build the production binary with `scripts/buildprod.sh`.
3. Authenticate to GCP with Workload Identity Federation.
4. Build and push `notely:latest` to Artifact Registry.
5. Run `scripts/migrateup.sh` against Turso.
6. Deploy the new image to Cloud Run.

It needs one GitHub Actions secret, `DATABASE_URL`, holding the full Turso
connection string including the token.

## Repo layout

```text
main.go                 server setup and routes
handler_*.go            HTTP handlers
middleware_auth.go      API key auth
internal/database/      sqlc-generated queries
internal/auth/          API key parsing
sql/schema/             goose migrations
sql/queries/            SQL that sqlc reads
scripts/                build and migrate helpers
static/index.html       the web page
.github/workflows/      CI and CD
```

## Status

The course is finished, so I deleted the Google Cloud resources to avoid
charges: the Cloud Run service, both Artifact Registry repositories, the Cloud
Build bucket, and the `notely_db_password` secret. The Turso database is still
there with my data.

That means the live URL no longer works and `cd.yml` will fail until the
infrastructure is recreated. To bring it back:

1. Create a Docker repository in Artifact Registry and a Cloud Run service.
2. Create the deploy service account and the Workload Identity pool and provider.
3. Store `DATABASE_URL` as a GitHub Actions secret and as a Secret Manager
   secret, and let the runtime service account read it.
4. Push to `main`.
