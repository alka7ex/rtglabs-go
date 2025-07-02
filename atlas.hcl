// atlas.hcl in your project root

# Define the migration directory
# This tells Atlas where your migration files are stored.
migrate {
  dir = "file:ent/migrate/migrations"
}

# Define the data source for your schema (your Ent schemas)
schema {
  url = "ent://ent/schema"
}

# Define an environment for development
# This environment sets the dev-url for diff commands and the URL for apply commands
env "dev" {
  # The URL to your actual development database file
  # This should be your BLUEPRINT_DB_URL value
  url = "sqlite:///home/farhienzahaikal/projects/rtglabs-go/db/test.db?_fk=true"

  # The URL for Atlas's "dev database" for diffing.
  # Using in-memory is great for SQLite.
  dev_url = "sqlite://file?mode=memory&_fk=1"

  # Reference the migrate directory defined above
  migration {
    dir = atlas.migrate.dir
  }
}

# You can define other environments like 'staging' or 'prod' here
# For example:
# env "prod" {
#   url = "mysql://user:pass@prod_db_host:3306/prod_db"
#   migration {
#     dir = atlas.migrate.dir
#   }
#   # No dev_url for prod environments, as you don't diff against prod directly
# }
