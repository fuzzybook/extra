ALTER TABLE broadcasters ADD COLUMN new_id UUID DEFAULT gen_random_uuid();
UPDATE broadcasters SET new_id = gen_random_uuid() WHERE new_id IS NULL;
ALTER TABLE broadcasters DROP CONSTRAINT broadcasters_pkey;
ALTER TABLE broadcasters DROP COLUMN id;
ALTER TABLE broadcasters RENAME COLUMN new_id TO id;
ALTER TABLE broadcasters ALTER COLUMN id SET NOT NULL;
ALTER TABLE broadcasters ADD PRIMARY KEY (id);

ALTER TABLE broadcasters ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE broadcasters ADD COLUMN used BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE broadcasters SET used = TRUE, enabled = TRUE;
