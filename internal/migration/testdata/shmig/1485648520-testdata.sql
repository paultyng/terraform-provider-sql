-- Migration: testdata
-- Created at: 2017-01-28 19:08:40
-- ====  UP  ====

BEGIN;

INSERT INTO shmig_test_table (code, name) VALUES ('QB' , 'Tom Brady');
INSERT INTO shmig_test_table (code, name) VALUES ('TE' , 'Ben Coates');
INSERT INTO shmig_test_table (code, name) VALUES ('CB' , 'Raymond Clayborn');
INSERT INTO shmig_test_table (code, name) VALUES ('G' ,  'John (Hog) Hannah');

COMMIT;

-- ==== DOWN ====

BEGIN;

DELETE FROM shmig_test_table;

COMMIT;