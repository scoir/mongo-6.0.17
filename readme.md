## background

a large `$group` expression seems to break and no longer group

## reproducing

A docker compose configuration is setup to create 6.0.16 and 6.0.17 instances of mongo.  The `tests` service (behind a profile) will execute the same set of inserts, query and validation steps against each instance.

**Expected outcome**
- 4 documents are created
- aggregation creates a single output document

### start the mongo instances

```
docker compose up -d
```

### execute the test script

```
docker compose run tests
```
