BEGIN;

Create type entry_direction as ENUM ('debit','creit');

-- An account is anything that holds money
create table accounts (
    id UUID primary key default gen_random_uuid (),
    -- system user
    owner_id UUID,
    name text not null check (name <> ''),
    currency text not null check (currency ~ '^[A-Z]{3}$'),
    normal_balance entry_direction not null,
    created_at timestamptz not null default now(),
)

COMMIT;