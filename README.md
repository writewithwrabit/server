# Wrabit Server

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
