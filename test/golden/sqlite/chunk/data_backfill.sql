-- migrasyon:chunk key (id)
-- migrasyon:statement 1
SELECT "users"."id" FROM "users" WHERE "users"."id" > ? ORDER BY "users"."id" LIMIT 1 OFFSET ?;

-- migrasyon:statement 2
UPDATE "users" SET "nickname" = 'anonymous' WHERE "users"."id" > ? AND "users"."id" <= ? AND "users"."nickname" IS NULL;

-- migrasyon:chunk key (tenant, id)
-- migrasyon:statement 1
SELECT "users"."tenant", "users"."id" FROM "users" WHERE ("users"."tenant", "users"."id") > (?, ?) ORDER BY "users"."tenant", "users"."id" LIMIT 1 OFFSET ?;

-- migrasyon:statement 2
UPDATE "users" SET "nickname" = 'anonymous' WHERE ("users"."tenant", "users"."id") > (?, ?) AND ("users"."tenant", "users"."id") <= (?, ?) AND "users"."nickname" IS NULL;
