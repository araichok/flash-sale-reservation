-- =========================
-- PRODUCTS
-- =========================
CREATE TABLE products (
                          id         BIGSERIAL PRIMARY KEY,
                          name       TEXT NOT NULL,
                          stock      INTEGER NOT NULL CHECK (stock >= 0),
                          created_at TIMESTAMP NOT NULL DEFAULT now()
);


-- =========================
-- RESERVATIONS
-- =========================
CREATE TABLE reservations (
                              id           BIGSERIAL PRIMARY KEY,
                              product_id   BIGINT NOT NULL REFERENCES products(id),
                              user_id      BIGINT NOT NULL,
                              status       TEXT NOT NULL,
                              expires_at   TIMESTAMP NOT NULL,
                              created_at   TIMESTAMP NOT NULL DEFAULT now()
);

-- ❗ Запрещаем более одного АКТИВНОГО резерва
-- для одного пользователя и одного товара
CREATE UNIQUE INDEX ux_active_reservation
    ON reservations (product_id, user_id)
    WHERE status = 'ACTIVE';

-- =========================
-- OUTBOX EVENTS
-- =========================
CREATE TABLE outbox_events (
                               id           BIGSERIAL PRIMARY KEY,
                               event_type   TEXT NOT NULL,
                               payload      JSONB NOT NULL,
                               created_at   TIMESTAMP NOT NULL DEFAULT now()
);
