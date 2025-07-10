CREATE TABLE secrets_access
(
    id        SERIAL PRIMARY KEY,
    secret_id INTEGER NOT NULL REFERENCES secrets (id) ON DELETE CASCADE,
    user_id   INTEGER REFERENCES users (id) ON DELETE CASCADE,
    role_id   INTEGER REFERENCES roles (id) ON DELETE CASCADE,
    CHECK (
        (user_id IS NOT NULL AND role_id IS NULL) OR
        (user_id IS NULL AND role_id IS NOT NULL)
        )
);