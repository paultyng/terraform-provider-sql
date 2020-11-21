-- Migration: create_table
-- Created at: 2017-01-28 22:39:14
-- ====  UP  ====

BEGIN;

CREATE TABLE
	shmig_test_table
(
	id	   integer
	, code     varchar(200)
	, name     varchar(200)
);

COMMIT;

-- ==== DOWN ====

BEGIN;

DROP TABLE
	shmig_test_table;

COMMIT;