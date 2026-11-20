BEGIN;

Create type entry_direction as ENUM ('debit','credit');

-- An account is anything that holds money
create table accounts (
    id UUID primary key default gen_random_uuid(),
    -- system user
    owner_id UUID,
    name text not null check (name <> ''),
    currency text not null check (currency ~ '^[A-Z]{3}$'),
    -- normal balance for knowing which side represent the increase (credit or debit)
    normal_balance entry_direction not null,
    created_at timestamptz not null default now(),
);

create index account_owner_id_idx on accounts (owner_id);

-- Journal book: for recording financial events
create table transactions (
    id UUID primary key default gen_random_uuid(),
    description text not null,
    created_at  timestamptz NOT NULL DEFAULT now()
);

create table entries (
    -- bigint: because its gonna used as an internal records and it is smaller than UUID, gives sequential IDs, and generally results in smaller indexes and efficient joins.
    id bigint  generated always as identity primary key,
    transaction_id uuid not null references transactions (id),
    account_id     uuid            NOT NULL,
    direction      entry_direction NOT NULL,
    amount         bigint          NOT NULL CHECK (amount > 0),
    currency       text            NOT NULL,
    created_at     timestamptz     NOT NULL DEFAULT now(),

    foreign key (account_id,currency) references accounts (id,currency)
);

create index entries_transaction_id_idx on entries (transaction_id);
create index entries_account_id_idx on entries (account_id);


create function check_transaction_balance() returns trigger 
LANGUAGE plpgsql AS $$
DECLARE
    tx_id uuid
begin
    if TG_TABLE_NAME = 'transactions' THEN
        tx_id := New.id;
    else 
        tx_id := New.transaction_id;
    end if;

    if not EXISTS (select 1 from entries where transaction_id = tx_id) then 
        raise EXCEPTION 'transaction %s as no entries', tx_id 
            using ERRCODE = 'check_violation';
    end if;

    if exists (
        select 1 from entries where transaction_id = tx_id
        group by currency
        having sum(case direction when 'debit' then amount else -amount end) <> 0
    ) then 
        raise EXCEPTION 'transaction %s is not balanced', tx_id 
            using ERRCODE = 'check_violation';
    end if;

    return null
end


create CONSTRAINT trigger transaction_balanced
after insert on transactions
DEFERRABLE INITIALLY DEFERRED
for each row execute function check_transaction_balance();


CREATE CONSTRAINT TRIGGER entries_balanced
after insert on entries
DEFERRABLE INITIALLY DEFERRED
for each row execute function check_transaction_balance();

create function forbid_ledger_changes() returns TRIGGER
LANGUAGE PLPGSQL AS $$
BEGIN
    RAISE EXCEPTION '% is append-only, post a reversing transaction instead', TG_TABLE_NAME
        USING ERRCODE = 'restrict_violation';
END


create TRIGGER transaction_append_only 
after DELETE or UPDATE on transactions
for each row EXECUTE forbid_ledger_changes();

CREATE TRIGGER entries_append_only
BEFORE UPDATE OR DELETE ON entries
FOR EACH ROW EXECUTE FUNCTION forbid_ledger_changes();

create view account_balances as
select a.id as account_id,a.name,a.currency,
COALESCE(sum(case when a.normal_balance = e.direction then e.amount else -e.amount end),0)::BIGINT as balance from accounts a left JOIN entries e on e.account_id = a.id
group by a.id

COMMIT;