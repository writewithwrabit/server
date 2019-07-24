CREATE TABLE users (
  id SERIAL,
  firebase_id VARCHAR,
  first_name VARCHAR,
  last_name VARCHAR,
  email VARCHAR,
  word_goal INT NOT NULL DEFAULT 1000
);

CREATE TABLE editors (
  id SERIAL,
  user_id VARCHAR,
  show_toolbar BOOLEAN NOT NULL DEFAULT true,
  show_prompt BOOLEAN NOT NULL DEFAULT false,
  show_counter BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE entries (
  id SERIAL,
  user_id VARCHAR,
  word_count INT,
  content TEXT
);

CREATE TABLE streaks (
  id SERIAL,
  user_id VARCHAR,
  day_count INT
);

INSERT INTO users (firebase_id, first_name, last_name, email, word_goal) VALUES ('0T3AWCd9mkdDFPeV0SDXqj3GRvZ2', 'Anthony', 'Morris', 'anthony@amorrissound.com', 1000);