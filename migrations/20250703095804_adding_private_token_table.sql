-- Create "private_tokens" table
CREATE TABLE `private_tokens` (
  `id` uuid NOT NULL,
  `token` text NOT NULL,
  `type` text NOT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  `user_private_token` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `private_tokens_users_private_token` FOREIGN KEY (`user_private_token`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "private_tokens_token_key" to table: "private_tokens"
CREATE UNIQUE INDEX `private_tokens_token_key` ON `private_tokens` (`token`);
