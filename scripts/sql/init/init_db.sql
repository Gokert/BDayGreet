DROP TABLE IF EXISTS profile CASCADE;
CREATE TABLE IF NOT EXISTS profile (
   id SERIAL NOT NULL PRIMARY KEY,
   login TEXT NOT NULL UNIQUE DEFAULT '',
   password bytea NOT NULL DEFAULT '',
   email TEXT NOT NULL DEFAULT '',
   birthday DATE NOT NULL
);

DROP TABLE IF EXISTS subscriber CASCADE;
CREATE TABLE IF NOT EXISTS subscriber(
    id_subscribe_from SERIAL NOT NULL REFERENCES profile(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,
    id_subscribe_to SERIAL NOT NULL REFERENCES profile(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,

    PRIMARY KEY(id_subscribe_to, id_subscribe_from)
);

CREATE INDEX idx_login ON profile(login);
CREATE INDEX idx_primary_key ON subscriber (id_subscribe_to, id_subscribe_from);


INSERT INTO profile(login, password, email, birthday) VALUES ('admin', '\xc7ad44cbad762a5da0a452f9e854fdc1e0e7a52a38015f23f3eab1d80b931dd472634dfac71cd34ebc35d16ab7fb8a90c81f975113d6c7538dc69dd8de9077ec', 'andreymyshlyaev9@gmail.com', '2005-01-01');