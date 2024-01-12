CREATE EXTENSION IF NOT EXISTS "uuid-ossp"
    VERSION "1.1";

CREATE TABLE wallets (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    balance FLOAT NOT NULL CHECK(balance >= 0.0),
    CONSTRAINT wallet_id_pkey PRIMARY KEY (id)
);

CREATE TABLE history (
    from_wallet uuid NOT NULL REFERENCES wallets(id),
    to_wallet uuid NOT NULL REFERENCES wallets(id),
    amount FLOAT NOT NULL CHECK(amount>0.0),
    time TIMESTAMPTZ NOT NULL DEFAULT now()
);