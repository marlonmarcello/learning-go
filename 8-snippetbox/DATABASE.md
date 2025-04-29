**Prerequisites**

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/) installed on your system.

**1. Setting up the PostgreSQL Container**

Open your terminal, navigate to the directory where you saved the file, and start the container in detached mode:

```bash
$ docker compose up -d
```

**2. Scaffolding the Table and Data**

Now you need to connect to the running PostgreSQL instance inside the container to create the necessary table and insert data. You'll connect as the `dbuser` that Docker created.

```bash
# Connect to the 'snippetbox' database as 'dbuser' inside the 'db' container
$ docker compose exec -it db psql -U dbuser -d snippetbox
```

You will be prompted for the password: `dbpass`
Now, inside the `psql` prompt, run the SQL commands to create the table and add the index. (Note: The database `snippetbox` already exists).

```sql
-- Create a `snippets` table.
CREATE TABLE snippets (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created TIMESTAMPTZ NOT NULL,
    expires TIMESTAMPTZ NOT NULL
);

-- Add an index on the created column.
CREATE INDEX idx_snippets_created ON snippets(created);
```

Insert the placeholder data using PostgreSQL's `CURRENT_TIMESTAMP` and interval syntax:

```sql
-- Add some dummy records.
INSERT INTO snippets (title, content, created, expires) VALUES (
    'An old silent pond',
    'An old silent pond...\nA frog jumps into the pond,\nsplash! Silence again.\n\n– Matsuo Bashō',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP + INTERVAL '365 days'
);

INSERT INTO snippets (title, content, created, expires) VALUES (
    'Over the wintry forest',
    'Over the wintry\nforest, winds howl in rage\nwith no leaves to blow.\n\n– Natsume Soseki',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP + INTERVAL '365 days'
);

INSERT INTO snippets (title, content, created, expires) VALUES (
    'First autumn morning',
    'First autumn morning\nthe mirror I stare into\nshows my father''s face.\n\n– Murakami Kijo',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP + INTERVAL '7 days'
);
```

**3. Creating a Less Privileged Application User**

For security, your web application should not connect as the `dbuser` (which is a superuser). Create a dedicated role with limited permissions. Stay in the same `psql` session (connected as `dbuser`).

```sql
-- Create a new role named 'web' that can log in and has a password.
-- Important: Replace 'your_strong_password' with a secure password for your application.
CREATE ROLE web WITH LOGIN PASSWORD 'pass';

-- Grant specific privileges on the snippets table to the 'web' role.
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE snippets TO web;

-- Grant usage permission on the sequence used by the 'id' column.
GRANT USAGE, SELECT ON SEQUENCE snippets_id_seq TO web;
```

Exit the `psql` session:

```sql
\q
```

**4. Test the Application User**

You can test the `web` user's permissions by connecting as that user.

Connect as `web` using `docker compose exec`:

```bash
# Connect to 'snippetbox' as 'web' inside the 'db' container, prompt for password
docker compose exec -it db psql -U web -d snippetbox -W
```

Enter the password you set for `web`: `pass`
Test the permissions:

```sql
-- This SELECT should work
SELECT id, title, expires FROM snippets;
```

```sh
  id |         title          |             expires
----+------------------------+-----------------------------------
  1 | An old silent pond     | 2026-04-28 18:05:00.123456+00  -- Example Timestamps
  2 | Over the wintry forest | 2026-04-28 18:05:00.123456+00
  3 | First autumn morning   | 2025-05-05 18:05:00.123456+00
(3 rows)
```

```sql
-- This DROP should fail (as expected)
DROP TABLE snippets;
```

```sh
ERROR:  permission denied for table snippets
```

Exit `psql`:

```sql
\q
```
