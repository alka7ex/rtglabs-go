-- Create "bodyweights" table
CREATE TABLE `bodyweights` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `weight` real NOT NULL,
  `unit` text NOT NULL,
  `user_bodyweights` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `bodyweights_users_bodyweights` FOREIGN KEY (`user_bodyweights`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "exercises" table
CREATE TABLE `exercises` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `name` text NOT NULL,
  PRIMARY KEY (`id`)
);
-- Create index "exercises_name_key" to table: "exercises"
CREATE UNIQUE INDEX `exercises_name_key` ON `exercises` (`name`);
-- Create "exercise_instances" table
CREATE TABLE `exercise_instances` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `exercise_exercise_instances` uuid NOT NULL,
  `workout_log_exercise_instances` uuid NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `exercise_instances_workout_logs_exercise_instances` FOREIGN KEY (`workout_log_exercise_instances`) REFERENCES `workout_logs` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `exercise_instances_exercises_exercise_instances` FOREIGN KEY (`exercise_exercise_instances`) REFERENCES `exercises` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "exercise_sets" table
CREATE TABLE `exercise_sets` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `weight` real NULL,
  `reps` integer NULL,
  `set_number` integer NOT NULL,
  `finished_at` datetime NULL,
  `status` integer NOT NULL DEFAULT 0,
  `exercise_exercise_sets` uuid NOT NULL,
  `exercise_instance_exercise_sets` uuid NULL,
  `workout_log_exercise_sets` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `exercise_sets_workout_logs_exercise_sets` FOREIGN KEY (`workout_log_exercise_sets`) REFERENCES `workout_logs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `exercise_sets_exercise_instances_exercise_sets` FOREIGN KEY (`exercise_instance_exercise_sets`) REFERENCES `exercise_instances` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `exercise_sets_exercises_exercise_sets` FOREIGN KEY (`exercise_exercise_sets`) REFERENCES `exercises` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "exerciseset_status" to table: "exercise_sets"
CREATE INDEX `exerciseset_status` ON `exercise_sets` (`status`);
-- Create "profiles" table
CREATE TABLE `profiles` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `units` integer NOT NULL,
  `age` integer NOT NULL,
  `height` real NOT NULL,
  `gender` integer NOT NULL,
  `weight` real NOT NULL,
  `user_profile` uuid NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `profiles_users_profile` FOREIGN KEY (`user_profile`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "profiles_user_profile_key" to table: "profiles"
CREATE UNIQUE INDEX `profiles_user_profile_key` ON `profiles` (`user_profile`);
-- Create "sessions" table
CREATE TABLE `sessions` (
  `id` uuid NOT NULL,
  `token` text NOT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  `user_sessions` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `sessions_users_sessions` FOREIGN KEY (`user_sessions`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sessions_token_key" to table: "sessions"
CREATE UNIQUE INDEX `sessions_token_key` ON `sessions` (`token`);
-- Create "users" table
CREATE TABLE `users` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `name` text NOT NULL,
  `email` text NOT NULL,
  `password` text NOT NULL,
  `email_verified_at` datetime NULL,
  PRIMARY KEY (`id`)
);
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX `users_email_key` ON `users` (`email`);
-- Create "workouts" table
CREATE TABLE `workouts` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `name` text NOT NULL,
  `user_workouts` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `workouts_users_workouts` FOREIGN KEY (`user_workouts`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "workout_exercises" table
CREATE TABLE `workout_exercises` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `order` integer NULL,
  `sets` integer NULL,
  `weight` real NULL,
  `reps` integer NULL,
  `exercise_workout_exercises` uuid NOT NULL,
  `exercise_instance_workout_exercises` uuid NULL,
  `workout_workout_exercises` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `workout_exercises_workouts_workout_exercises` FOREIGN KEY (`workout_workout_exercises`) REFERENCES `workouts` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `workout_exercises_exercise_instances_workout_exercises` FOREIGN KEY (`exercise_instance_workout_exercises`) REFERENCES `exercise_instances` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `workout_exercises_exercises_workout_exercises` FOREIGN KEY (`exercise_workout_exercises`) REFERENCES `exercises` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "workout_logs" table
CREATE TABLE `workout_logs` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `started_at` datetime NULL,
  `finished_at` datetime NULL,
  `status` integer NOT NULL DEFAULT 0,
  `total_active_duration_seconds` integer NOT NULL DEFAULT 0,
  `total_pause_duration_seconds` integer NOT NULL DEFAULT 0,
  `user_workout_logs` uuid NOT NULL,
  `workout_workout_logs` uuid NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `workout_logs_workouts_workout_logs` FOREIGN KEY (`workout_workout_logs`) REFERENCES `workouts` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `workout_logs_users_workout_logs` FOREIGN KEY (`user_workout_logs`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "workoutlog_status" to table: "workout_logs"
CREATE INDEX `workoutlog_status` ON `workout_logs` (`status`);
