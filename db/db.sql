CREATE EXTENSION IF NOT EXISTS CITEXT;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS forums CASCADE;
DROP TABLE IF EXISTS threads CASCADE;
DROP TABLE IF EXISTS Thread_vote CASCADE;
DROP TABLE IF EXISTS posts CASCADE;
DROP TABLE IF EXISTS Forum_user CASCADE;

CREATE UNLOGGED TABLE IF NOT EXISTS users (
    id SERIAL UNIQUE NOT NULL,
    nickname CITEXT NOT NULL PRIMARY KEY,
    email    CITEXT NOT NULL UNIQUE,
    fullname CITEXT NOT NULL,
    about    TEXT   NOT NULL
);

CREATE  INDEX index_users_id ON users (id);
CREATE INDEX index_users_nickname ON users (nickname);
CREATE INDEX index_users_email ON users (email);


CREATE UNLOGGED TABLE IF NOT EXISTS forums (
    id           SERIAL,
    slug         CITEXT PRIMARY KEY,
    post_count   INT    NOT NULL DEFAULT 0,
    thread_count INT       NOT NULL DEFAULT 0,
    title        TEXT      NOT NULL,
    user_nickname  CITEXT      NOT NULL
);

CREATE INDEX index_forums_id_hash ON forums USING HASH (id);
CREATE INDEX index_forums_slug_hash ON forums USING HASH (slug);
CREATE INDEX index_forums_users_foreign ON forums (user_nickname);


CREATE UNLOGGED TABLE IF NOT EXISTS threads (
    id         SERIAL PRIMARY KEY ,
    author    CITEXT        NOT NULL REFERENCES users(nickname),
    created   TIMESTAMP WITH TIME ZONE DEFAULT now(),
    forum     CITEXT        NOT NULL REFERENCES forums(slug),
    msg       TEXT        NOT NULL,
    slug      CITEXT      UNIQUE,
    title     TEXT        NOT NULL,
    votes     INT         NOT NULL DEFAULT 0,
    FOREIGN KEY (forum) REFERENCES Forums (slug) ON DELETE CASCADE,
    FOREIGN KEY (author) REFERENCES Users (nickname) ON DELETE CASCADE
);

CREATE INDEX index_threads_slug_hash ON threads USING HASH (slug);
CREATE INDEX index_threads_id ON threads (id);


CREATE OR REPLACE FUNCTION threads_forum_counter()
    RETURNS TRIGGER AS $threads_forum_counter$
BEGIN
UPDATE forums
SET thread_count = thread_count + 1
WHERE slug = NEW.forum;
RETURN NULL;
END;
$threads_forum_counter$  LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS threads_forum_counter ON threads;
CREATE TRIGGER threads_forum_counter AFTER INSERT ON threads FOR EACH ROW EXECUTE PROCEDURE threads_forum_counter();


CREATE UNLOGGED TABLE posts (
    id SERIAL PRIMARY KEY ,
    path INTEGER[],
    author CITEXT NOT NULL REFERENCES users(nickname),
    created TIMESTAMP WITH TIME ZONE DEFAULT now(),
    isEdited BOOLEAN DEFAULT FALSE,
    msg      TEXT  NOT NULL,
    parent   INTEGER,
    forum CITEXT NOT NULL,
    thread INTEGER NOT NULL
);

CREATE INDEX index_posts_id on posts (id);
CREATE INDEX index_posts_thread_id on posts (thread, id);
CREATE INDEX index_posts_path1_path on posts ((path[1]), path);


CREATE UNLOGGED TABLE Forum_user (
    forum_slug CITEXT NOT NULL,
    nickname CITEXT NOT NULL,
    UNIQUE (forum_slug, nickname),
    FOREIGN KEY (nickname) REFERENCES Users (nickname)
);


CREATE OR REPLACE FUNCTION add_forum_user()
    RETURNS TRIGGER AS
$add_forum_user$
BEGIN
INSERT INTO forum_user (nickname, forum_slug)
VALUES (new.author, new.forum)
    ON CONFLICT DO NOTHING;
RETURN new;
END;
$add_forum_user$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION add_forum_user_thread()
    RETURNS TRIGGER AS
$add_forum_user_thread$
BEGIN
INSERT INTO forum_user (nickname, forum_slug)
VALUES (new.author, new.forum)
    ON CONFLICT DO NOTHING;
RETURN new;
END;
$add_forum_user_thread$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION set_edited() RETURNS TRIGGER AS $set_edited$
BEGIN
    IF (NEW.msg = OLD.msg)
    THEN RETURN NULL;
END IF;
UPDATE posts SET isEdited = TRUE
WHERE id=NEW.id;
RETURN NULL;
END;
$set_edited$  LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_edited ON posts;
CREATE TRIGGER set_edited AFTER UPDATE ON posts FOR EACH ROW EXECUTE PROCEDURE set_edited();

CREATE OR REPLACE FUNCTION check_edited(pid INT, message TEXT)
    RETURNS BOOLEAN AS $check_edited$
BEGIN
    IF ((SELECT posts.msg FROM posts WHERE id=pid) = message)
    THEN RETURN FALSE;
END IF;
RETURN TRUE;
END;
$check_edited$ LANGUAGE plpgsql;


CREATE UNLOGGED TABLE IF NOT EXISTS Thread_vote (
    nickname   CITEXT REFERENCES users(nickname)   NOT NULL,
    thread_id INT REFERENCES threads(id)          NOT NULL,
    vote     INT                                 NOT NULL
);


CREATE OR REPLACE FUNCTION vote_insert()
    RETURNS TRIGGER AS $vote_insert$
BEGIN
UPDATE threads
SET votes = votes + NEW.vote
WHERE id = NEW.thread_id;
RETURN NULL;
END;
$vote_insert$  LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS vote_insert ON Thread_vote;
CREATE TRIGGER vote_insert AFTER INSERT ON Thread_vote FOR EACH ROW EXECUTE PROCEDURE vote_insert();


CREATE OR REPLACE FUNCTION vote_update() RETURNS TRIGGER AS $vote_update$
BEGIN
    IF OLD.vote = NEW.vote
    THEN
        RETURN NULL;
END IF;
UPDATE threads
SET
    votes = votes + CASE WHEN NEW.vote = -1
                             THEN -2
                         ELSE 2 END
WHERE id = NEW.thread_id;
RETURN NULL;
END;
$vote_update$ LANGUAGE  plpgsql;

DROP TRIGGER IF EXISTS vote_update ON Thread_vote;
CREATE TRIGGER vote_update AFTER UPDATE ON Thread_vote FOR EACH ROW EXECUTE PROCEDURE vote_update();

CREATE OR REPLACE FUNCTION set_post_path()
    RETURNS TRIGGER AS
$set_post_path$
BEGIN
    new.path = (SELECT path FROM posts WHERE id = new.parent) || new.id;
UPDATE forums SET post_count = post_count + 1 WHERE slug = new.forum;
RETURN new;
END;
$set_post_path$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_forum_threads()
    RETURNS TRIGGER AS
$update_forum_threads$
BEGIN
UPDATE forums SET thread_count = thread_count + 1 WHERE slug = new.forum;
RETURN new;
END;
$update_forum_threads$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS path_trigger ON posts;
CREATE TRIGGER path_trigger
    BEFORE INSERT
    ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE path();

DROP TRIGGER IF EXISTS add_forum_user_new_post ON posts;
CREATE TRIGGER add_forum_user_new_post
    AFTER INSERT
    ON Posts
    FOR EACH ROW
    EXECUTE PROCEDURE add_forum_user();

DROP TRIGGER IF EXISTS add_forum_user_new_thread ON threads;
CREATE TRIGGER add_forum_user_new_thread
    AFTER INSERT
    ON threads
    FOR EACH ROW
    EXECUTE PROCEDURE add_forum_user_thread();

DROP TRIGGER IF EXISTS set_post_path ON Posts;
CREATE TRIGGER set_post_path
    BEFORE INSERT
    ON Posts
    FOR EACH ROW
    EXECUTE PROCEDURE set_post_path();
