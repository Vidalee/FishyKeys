CREATE TABLE IF NOT EXISTS secrets
(
    id
    SERIAL
    PRIMARY
    KEY,
    path
    VARCHAR
(
    255
) NOT NULL UNIQUE,
    encrypted_encryption_key VARCHAR
(
    128
) NOT NULL,
    encrypted_value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );