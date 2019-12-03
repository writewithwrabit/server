# Wrabit Server

## Getting Started

### Requirements

For now the following are required:

- Stripe (for payments)
- Firebase (for auth)
- Mailgun (for sending emails)

I will do my best to work on removing these requirements in dev but I don't know if/how/when that will happen.

### Environment

```env
// Sets your environment to dev which turns off some handy things
// like the Sqreen agent. It also turns on the GraphQL playground
NODE_ENV=dev

// The username to login to postgres wth
CLOUDSQL_USER=postgres

// The password for the postgres user
CLOUDSQL_PASSWORD=allthesecurity

// For local dev this is the name of the Docker DB container (the host)
CLOUDSQL_CONNECTION_NAME=database

// Database to connect to
CLOUDSQL_DATABASE_NAME=wrabit

// Client secret used to connect with Firebase
GOOGLE_APPLICATION_CREDENTIALS=client-secret.json

// Used to interact with the Stripe platform
STRIPE_KEY=XXXXXXXXXXXXXXXXXXXX

// Used to send email through mailgun
MAILGUN_KEY=XXXXXXXXXXXXXXXXXXXX

// Used to encrypt user data
ENCRYPTION_KEY=thisencryptsuserdatainthedatabase
```

### Setup

1. Create required accounts (see above)
2. Copy `.env.example` to `.env` and fill out the fields
3. Modify last line of `wrabit.sql` to have a user for testing (or manually create an account)
4. Run `docker-compose up`

## Generate GraphQL Schema

1. Make changes to `schema.graphql`
2. Run `go generate ./...` from the root directory

## Managing SQL Schema

The schema is currently managed by one SQL file (`wrabit.sql`). Once the database becomes larger, we will be forced to solve the schema management problem. Until then...

### Connect to GCP SQL

1. Run the `gcloud` SQL connect command (this will whitelist your IP for 5 minutes)

    ```bash
    gcloud sql connect wrabit-postgres
    ```

2. Enter password for database (check `.prod.env` file)

3. Switch to the `wrabit` database

    ```psql
    \c wrabit
    ```

## Encrypting Secrets

Secrets are currently stored in a local `.env` file. In order to get them onto the CI/CD pipeline, we need [to use Travis' encrypt tool](https://docs.travis-ci.com/user/encryption-keys/).

1. Zip the secrets you want into a `secrets.tar` file (which is git ignored)

    ```bash
    tar cvf secrets.tar .prod.env client-secret.json sqreen.yaml
    ```

2. Encrypt the zip with the Travis encrypt tool

    ```bash
    travis encrypt-file secrets.tar --add --com
    ```
