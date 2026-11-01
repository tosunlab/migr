-- migrasyon:chunk key (id)
-- migrasyon:statement 1
SELECT `id` FROM `users` WHERE `id` > ? ORDER BY `id` LIMIT 1 OFFSET ?;

-- migrasyon:statement 2
UPDATE `users` SET `email_key` = LOWER(CAST(`email` AS CHAR)) WHERE `id` > ? AND `id` <= ?;

-- migrasyon:chunk key (tenant, id)
-- migrasyon:statement 1
SELECT `tenant`, `id` FROM `users` WHERE `tenant` >= ? AND (`tenant` > ? OR (`id` > ?)) ORDER BY `tenant`, `id` LIMIT 1 OFFSET ?;

-- migrasyon:statement 2
UPDATE `users` SET `email_key` = LOWER(CAST(`email` AS CHAR)) WHERE `tenant` >= ? AND (`tenant` > ? OR (`id` > ?)) AND `tenant` <= ? AND (`tenant` < ? OR (`id` <= ?));
