BEGIN;

Create type entry_direction as ENUM ('debit','creit');

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


COMMIT;