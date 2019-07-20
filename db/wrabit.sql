CREATE TABLE users (
  id varchar,
  first_name varchar,
  last_name varchar,
  email varchar,
  word_goal int DEFAULT 1000
);

CREATE TABLE editors (
  id SERIAL,
  user_id varchar,
  show_toolbar boolean,
  show_prompt boolean,
  show_counter boolean
);

CREATE TABLE entries (
  id SERIAL,
  user_id varchar,
  word_count int,
  content text
);

CREATE TABLE streaks (
  id SERIAL,
  user_id varchar,
  day_count int
);

INSERT INTO users (id, first_name, last_name, email, word_goal) VALUES ('0T3AWCd9mkdDFPeV0SDXqj3GRvZ2', 'Anthony', 'Morris', 'anthony@amorrissound.com', 1000);