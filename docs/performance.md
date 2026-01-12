## Possible performance optimizations

- Use `jsonb`, meaning:
    - Data in PGSQL must be re-formatted to `jsonb` format.
    - Backend remains as-is.
    - Queries change to adhere to a new DB format. 

- Create indices on `Candidates` and `Positions` tables.
- Create in-memory storage for quickly checking a match without expensive DB queries.
