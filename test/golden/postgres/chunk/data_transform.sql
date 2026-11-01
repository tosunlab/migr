-- migrasyon:chunk key (id)
-- migrasyon:statement 1
SELECT "id"::text FROM "users" WHERE "id" > $1 ORDER BY "users"."id" LIMIT 1 OFFSET $2;

-- migrasyon:statement 2
UPDATE "users" SET "email_key" = CAST(lower(CAST("email" AS text)) AS varchar(255)) WHERE "id" > $1 AND "id" <= $2;

-- migrasyon:chunk key (tenant, id)
-- migrasyon:statement 1
SELECT "tenant"::text, "id"::text FROM "users" WHERE ("tenant", "id") > ($1, $2) ORDER BY "users"."tenant", "users"."id" LIMIT 1 OFFSET $3;

-- migrasyon:statement 2
UPDATE "users" SET "email_key" = CAST(lower(CAST("email" AS text)) AS varchar(255)) WHERE ("tenant", "id") > ($1, $2) AND ("tenant", "id") <= ($3, $4);
