# stats-for-orders

HTTP service that can save and return information about orders.
Information is stored in ClickHouse.
Examples of all HTTP handlers can be found in the spec.txt file.
Additionally, there are end-to-end tests for the entire service in the "tests" directory.

The service can be started using Docker Compose:
```
docker compose up --build
```
After it starts, database tables are migrated if necessary.
Migrations are stored in the "migrations" directory.
