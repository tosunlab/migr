import { literal, replaceColumn, t } from "@migrasyon/sdk";

/** The contract of renaming accounts.handle, irreversible after a lowercasing backfill. */
export default replaceColumn("accounts", "handle", t.string(64))
  .wasNull()
  .wasDefault(literal("anon"))
  .renameTo("login")
  .default(literal("guest"))
  .through("lower")
  .contract();
