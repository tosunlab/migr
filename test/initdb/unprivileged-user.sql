-- A second user holding exactly what a migrasyon deployment holds: its own database, and no PROCESS.
--
-- grant-process.sql gives the test user PROCESS so the lock-impact suites can watch InnoDB's
-- TABLE_ID. That also makes every catalog readable, which hides the one question M13-T1's Assess
-- has to answer honestly: what can a session lowering a verdict read without that grant? This user
-- is how the suite asks it.
CREATE USER 'migrasyon_plain'@'%' IDENTIFIED BY 'migrasyon_plain';
GRANT ALL PRIVILEGES ON migrasyon.* TO 'migrasyon_plain'@'%';
FLUSH PRIVILEGES;
