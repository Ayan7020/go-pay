BEGIN;

CREATE INDEX entries_account_id_idx ON entries (account_id);
DROP INDEX entries_account_id_id_idx;

ALTER TABLE accounts DROP COLUMN kind;
DROP TYPE account_kind;
COMMIT;