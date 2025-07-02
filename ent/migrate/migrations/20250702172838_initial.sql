-- Create "bodyweights" table
CREATE TABLE `bodyweights` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `weight` real NOT NULL,
  `unit` text NOT NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `bodyweights_users_bodyweights` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
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
  `workout_log_id` uuid NULL,
  `exercise_id` uuid NOT NULL,
  `exercise_exercise_instances` uuid NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `exercise_instances_exercises_exercise_instances` FOREIGN KEY (`exercise_exercise_instances`) REFERENCES `exercises` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "exerciseinstance_workout_log_id" to table: "exercise_instances"
CREATE INDEX `exerciseinstance_workout_log_id` ON `exercise_instances` (`workout_log_id`);
-- Create index "exerciseinstance_exercise_id" to table: "exercise_instances"
CREATE INDEX `exerciseinstance_exercise_id` ON `exercise_instances` (`exercise_id`);
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
  `user_id` uuid NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `profiles_users_profile` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "profiles_user_id_key" to table: "profiles"
CREATE UNIQUE INDEX `profiles_user_id_key` ON `profiles` (`user_id`);
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
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `workouts_users_workouts` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "workout_user_id" to table: "workouts"
CREATE INDEX `workout_user_id` ON `workouts` (`user_id`);
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
  `exercise_id` uuid NOT NULL,
  `exercise_instance_id` uuid NULL,
  `workout_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `workout_exercises_workouts_workout_exercises` FOREIGN KEY (`workout_id`) REFERENCES `workouts` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `workout_exercises_exercise_instances_workout_exercises` FOREIGN KEY (`exercise_instance_id`) REFERENCES `exercise_instances` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `workout_exercises_exercises_workout_exercises` FOREIGN KEY (`exercise_id`) REFERENCES `exercises` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "workoutexercise_workout_id_exercise_id" to table: "workout_exercises"
CREATE UNIQUE INDEX `workoutexercise_workout_id_exercise_id` ON `workout_exercises` (`workout_id`, `exercise_id`);
-- Create index "workoutexercise_workout_id" to table: "workout_exercises"
CREATE INDEX `workoutexercise_workout_id` ON `workout_exercises` (`workout_id`);
-- Create index "workoutexercise_exercise_id" to table: "workout_exercises"
CREATE INDEX `workoutexercise_exercise_id` ON `workout_exercises` (`exercise_id`);
-- Create index "workoutexercise_exercise_instance_id" to table: "workout_exercises"
CREATE INDEX `workoutexercise_exercise_instance_id` ON `workout_exercises` (`exercise_instance_id`);
