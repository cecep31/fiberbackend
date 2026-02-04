-- Holding types lookup table (referenced by holdings.holding_type_id)
CREATE TABLE IF NOT EXISTS holding_types (
  id smallint PRIMARY KEY,
  name text
);

-- Holdings table (matches Drizzle schema)
CREATE TABLE IF NOT EXISTS holdings (
  id bigserial PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name text NOT NULL,
  platform text NOT NULL,
  holding_type_id smallint NOT NULL REFERENCES holding_types(id) ON DELETE RESTRICT,
  currency char(3) NOT NULL,
  invested_amount numeric(18, 2) NOT NULL DEFAULT 0,
  current_value numeric(18, 2) NOT NULL DEFAULT 0,
  units numeric(24, 10),
  avg_buy_price numeric(18, 8),
  current_price numeric(18, 8),
  last_updated timestamptz,
  notes text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  month integer NOT NULL DEFAULT 1,
  year integer NOT NULL DEFAULT 2025,
  CONSTRAINT chk_holdings_positive_amounts CHECK (invested_amount >= 0 AND current_value >= 0),
  CONSTRAINT chk_holdings_valid_month CHECK (month >= 1 AND month <= 12),
  CONSTRAINT chk_holdings_valid_year CHECK (year >= 2000)
);

CREATE INDEX IF NOT EXISTS idx_holdings_user ON holdings (user_id);
CREATE INDEX IF NOT EXISTS idx_holdings_holding_type_id ON holdings (holding_type_id);
CREATE INDEX IF NOT EXISTS idx_holdings_month_year ON holdings (year, month);
CREATE INDEX IF NOT EXISTS idx_holdings_user_month_year ON holdings (user_id, year, month);
