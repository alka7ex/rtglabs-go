-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "name" character varying NOT NULL,
  "email" character varying NOT NULL,
  "password" character varying NOT NULL,
  "email_verified_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "public"."users" ("email");
-- Create "bodyweights" table
CREATE TABLE "public"."bodyweights" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "weight" double precision NOT NULL,
  "unit" character varying NOT NULL,
  "user_bodyweights" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "bodyweights_users_bodyweights" FOREIGN KEY ("user_bodyweights") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "exercises" table
CREATE TABLE "public"."exercises" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "name" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "exercises_name_key" to table: "exercises"
CREATE UNIQUE INDEX "exercises_name_key" ON "public"."exercises" ("name");
-- Create "workouts" table
CREATE TABLE "public"."workouts" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "name" character varying NOT NULL,
  "user_workouts" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "workouts_users_workouts" FOREIGN KEY ("user_workouts") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "workout_logs" table
CREATE TABLE "public"."workout_logs" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "started_at" timestamptz NULL,
  "finished_at" timestamptz NULL,
  "status" bigint NOT NULL DEFAULT 0,
  "total_active_duration_seconds" bigint NOT NULL DEFAULT 0,
  "total_pause_duration_seconds" bigint NOT NULL DEFAULT 0,
  "user_workout_logs" uuid NOT NULL,
  "workout_workout_logs" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "workout_logs_users_workout_logs" FOREIGN KEY ("user_workout_logs") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "workout_logs_workouts_workout_logs" FOREIGN KEY ("workout_workout_logs") REFERENCES "public"."workouts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "workoutlog_status" to table: "workout_logs"
CREATE INDEX "workoutlog_status" ON "public"."workout_logs" ("status");
-- Create "exercise_instances" table
CREATE TABLE "public"."exercise_instances" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "exercise_exercise_instances" uuid NOT NULL,
  "workout_log_exercise_instances" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "exercise_instances_exercises_exercise_instances" FOREIGN KEY ("exercise_exercise_instances") REFERENCES "public"."exercises" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "exercise_instances_workout_logs_exercise_instances" FOREIGN KEY ("workout_log_exercise_instances") REFERENCES "public"."workout_logs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "exercise_sets" table
CREATE TABLE "public"."exercise_sets" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "weight" numeric(8,2) NULL,
  "reps" bigint NULL,
  "set_number" bigint NOT NULL,
  "finished_at" timestamptz NULL,
  "status" bigint NOT NULL DEFAULT 0,
  "exercise_exercise_sets" uuid NOT NULL,
  "exercise_instance_exercise_sets" uuid NULL,
  "workout_log_exercise_sets" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "exercise_sets_exercise_instances_exercise_sets" FOREIGN KEY ("exercise_instance_exercise_sets") REFERENCES "public"."exercise_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "exercise_sets_exercises_exercise_sets" FOREIGN KEY ("exercise_exercise_sets") REFERENCES "public"."exercises" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "exercise_sets_workout_logs_exercise_sets" FOREIGN KEY ("workout_log_exercise_sets") REFERENCES "public"."workout_logs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "exerciseset_status" to table: "exercise_sets"
CREATE INDEX "exerciseset_status" ON "public"."exercise_sets" ("status");
-- Create "private_tokens" table
CREATE TABLE "public"."private_tokens" (
  "id" uuid NOT NULL,
  "token" character varying NOT NULL,
  "type" character varying NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  "user_private_token" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "private_tokens_users_private_token" FOREIGN KEY ("user_private_token") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "private_tokens_token_key" to table: "private_tokens"
CREATE UNIQUE INDEX "private_tokens_token_key" ON "public"."private_tokens" ("token");
-- Create "profiles" table
CREATE TABLE "public"."profiles" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "units" bigint NOT NULL,
  "age" bigint NOT NULL,
  "height" numeric(10,2) NOT NULL,
  "gender" bigint NOT NULL,
  "weight" numeric(10,2) NOT NULL,
  "user_profile" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "profiles_users_profile" FOREIGN KEY ("user_profile") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "profiles_user_profile_key" to table: "profiles"
CREATE UNIQUE INDEX "profiles_user_profile_key" ON "public"."profiles" ("user_profile");
-- Create "sessions" table
CREATE TABLE "public"."sessions" (
  "id" uuid NOT NULL,
  "token" character varying NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  "user_sessions" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sessions_users_sessions" FOREIGN KEY ("user_sessions") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sessions_token_key" to table: "sessions"
CREATE UNIQUE INDEX "sessions_token_key" ON "public"."sessions" ("token");
-- Create "workout_exercises" table
CREATE TABLE "public"."workout_exercises" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "order" bigint NULL,
  "sets" bigint NULL,
  "weight" double precision NULL,
  "reps" bigint NULL,
  "exercise_workout_exercises" uuid NOT NULL,
  "exercise_instance_workout_exercises" uuid NULL,
  "workout_workout_exercises" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "workout_exercises_exercise_instances_workout_exercises" FOREIGN KEY ("exercise_instance_workout_exercises") REFERENCES "public"."exercise_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "workout_exercises_exercises_workout_exercises" FOREIGN KEY ("exercise_workout_exercises") REFERENCES "public"."exercises" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "workout_exercises_workouts_workout_exercises" FOREIGN KEY ("workout_workout_exercises") REFERENCES "public"."workouts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
