-- migrasyon:up
CREATE TABLE authors (
  id   INTEGER PRIMARY KEY,
  name TEXT NOT NULL
);

-- migrasyon:down
DROP TABLE authors;
