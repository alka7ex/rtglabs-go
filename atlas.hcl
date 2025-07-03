
env "local" {
  # Dev SQLite used for diffing
  dev = "sqlite://file?mode=memory&cache=shared&_fk=1"

  # Actual database
  url = "sqlite://db/test.db"

  # Where to store the generated migration files
  migration {
    dir = "file://migrations"
  }

  # Your Ent schema as the source of truth
  src = "ent://ent/schema"
}

