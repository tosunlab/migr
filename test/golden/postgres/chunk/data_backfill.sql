-- migrasyon:chunk key (id)
-- migrasyon:statement 1
SELECT "id"::text FROM "users" WHERE "id" > $1 ORDER BY "users"."id" LIMIT 1 OFFSET $2;

-- migrasyon:statement 2
UPDATE "users" SET "nickname" = 'anonymous' WHERE "id" > $1 AND "id" <= $2 AND "nickname" IS NULL;

-- migrasyon:chunk key (tenant, id)
-- migrasyon:statement 1
SELECT "tenant"::text, "id"::text FROM "users" WHERE ("tenant", "id") > ($1, $2) ORDER BY "users"."tenant", "users"."id" LIMIT 1 OFFSET $3;

-- migrasyon:statement 2
UPDATE "users" SET "nickname" = 'anonymous' WHERE ("tenant", "id") > ($1, $2) AND ("tenant", "id") <= ($3, $4) AND "nickname" IS NULL;
