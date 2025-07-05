env "prod" {
  # The actual production database URL
  # IMPORTANT: This should be an environment variable or a secret in a real production setup!
  # NEVER hardcode credentials or full URLs in your config file.
  # For example: POSTGRES_URL="postgres://user:password@host:port/database?sslmode=disable"
  # This is the line you need to change:
  url="postgresql://postgres.zvvuxfudwcsztctduylx:fnAbvE0XxaFdWu@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres?search_path=public"
  #                                                                               ^^^^^^^^^^^^^^^^^^ Add this part

  # Dev database for planning/linting migrations against a real PostgreSQL instance
  # This is usually a separate, ephemeral database or a development instance
  dev = "docker://postgres/16/atlas_dev_db" # Example using Docker for a dev container

  revisions_schema = "public" # Keep this line as well for explicit configuration
  # Where to store the generated migration files (can be shared with local)
  migration {
    dir = "file://migrations"
  }

  # Your Ent schema as the source of truth
  src = "ent://ent/schema"
}
