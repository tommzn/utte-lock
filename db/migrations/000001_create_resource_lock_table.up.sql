CREATE TABLE resource_locks (
  resource_id   VARCHAR (128)   PRIMARY KEY, 
  client_id     VARCHAR (128)   NOT NULL, 
  expiry        INTEGER         NOT NULL DEFAULT 0, 
  sequence_no   SERIAL,
  created_at    TIMESTAMP       DEFAULT NOW() 
);