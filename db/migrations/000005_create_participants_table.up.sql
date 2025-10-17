CREATE TABLE public.participants (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type        VARCHAR(20) NOT NULL CHECK (type IN ('wallet', 'internal')),
    ref_id      INTEGER NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(type, ref_id) -- supaya 1 wallet/internal tidak duplikat
);