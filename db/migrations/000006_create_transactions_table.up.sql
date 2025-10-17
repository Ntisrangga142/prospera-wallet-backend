CREATE TYPE type_transaction AS ENUM ('transfer', 'top_up');

CREATE TABLE public.transactions (
    id                  INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type                type_transaction NOT NULL,
    amount              INTEGER NOT NULL,
    total               INTEGER NOT NULL,
    note                TEXT,
    id_sender           INTEGER NOT NULL,
    id_receiver         INTEGER NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_sender      TIMESTAMP,
    deleted_receiver    TIMESTAMP,
    CONSTRAINT fk_tx_sender FOREIGN KEY (id_sender) REFERENCES participants(id),
    CONSTRAINT fk_tx_receiver FOREIGN KEY (id_receiver) REFERENCES participants(id)
);
