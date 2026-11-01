CREATE TABLE items (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(150) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  stock       INTEGER NOT NULL DEFAULT 0 CONSTRAINT chk_items_stock_non_negative CHECK (stock >= 0),
  price       BIGINT  NOT NULL CONSTRAINT chk_items_price_non_negative CHECK (price >= 0),
  created_by  UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_items_created_at_id ON items (created_at DESC, id DESC);
CREATE INDEX idx_items_name_lower ON items (lower(name));

