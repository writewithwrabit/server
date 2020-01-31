CREATE TABLE users (
  id SERIAL,
  firebase_id VARCHAR,
  stripe_id VARCHAR,
  stripe_subscription_id VARCHAR,
  first_name VARCHAR,
  last_name VARCHAR,
  email VARCHAR,
  word_goal INT NOT NULL DEFAULT 1000,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE editors (
  id SERIAL,
  user_id VARCHAR,
  show_toolbar BOOLEAN NOT NULL DEFAULT true,
  show_prompt BOOLEAN NOT NULL DEFAULT false,
  show_counter BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE entries (
  id SERIAL,
  user_id VARCHAR,
  word_count INT,
  content TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  goal_hit BOOLEAN DEFAULT false
);

CREATE TABLE streaks (
  id SERIAL,
  user_id VARCHAR,
  day_count INT,
  last_entry_id INT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE donations (
  id SERIAL,
  user_id VARCHAR,
  amount INT,
  paid BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  entry_id VARCHAR
);

CREATE OR REPLACE FUNCTION trigger_updated()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER updated
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE trigger_updated();

CREATE TRIGGER updated
BEFORE UPDATE ON editors
FOR EACH ROW
EXECUTE PROCEDURE trigger_updated();

CREATE TRIGGER updated
BEFORE UPDATE ON entries
FOR EACH ROW
EXECUTE PROCEDURE trigger_updated();

CREATE TRIGGER updated
BEFORE UPDATE ON streaks
FOR EACH ROW
EXECUTE PROCEDURE trigger_updated();

CREATE TRIGGER updated
BEFORE UPDATE ON donations
FOR EACH ROW
EXECUTE PROCEDURE trigger_updated();

INSERT INTO users (firebase_id, stripe_id, stripe_subscription_id, first_name, last_name, email, word_goal) VALUES ('6uP1r7qI8ZaYetQcGG6GYYYB2Em2', 'cus_GIHI1V0ryeznB2', 'sub_GIHImr4be4B275', 'Test', 'Account', 'testing@writewithwrabit.com', 1000);
