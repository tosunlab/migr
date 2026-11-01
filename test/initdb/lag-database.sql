-- The database the replica-lag tests write through a MySQL replica (M13-T8).
--
-- The compose replica replicates this database and no other (--replicate-do-db), because every
-- other suite creates and drops tables in `migrasyon` all day, and a replica that met a statement about a
-- table it never saw would stop its SQL thread -- which is exactly the state those tests read as
-- unknown lag. It is created with the binary log off, so the replica, which also creates it, never
-- receives a second CREATE DATABASE. MariaDB runs this file too and simply has the database spare.
SET sql_log_bin = 0;
CREATE DATABASE IF NOT EXISTS migrasyon_lag;
GRANT ALL ON migrasyon_lag.* TO 'migrasyon'@'%';
FLUSH PRIVILEGES;
