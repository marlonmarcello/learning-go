The `internal` directory will contain the ancillary _non-application-specific_ code used in the project. We’ll use it to hold potentially reusable code like validation helpers and the SQL database models for the project.

It’s important to point out that the directory name `internal` carries a special meaning and behavior in Go: any packages which live under this directory can only be imported by code _inside the parent of the `internal` directory_. In our case, this means that any packages which live in `internal` can only be imported by code inside our `snippetbox` project directory.
