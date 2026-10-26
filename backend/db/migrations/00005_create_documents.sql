-- +goose Up
CREATE TABLE documents (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title       text NOT NULL,
    content_md  text NOT NULL,
    category    text NOT NULL
                CHECK (category IN ('procedimiento', 'runbook', 'guia', 'configuracion', 'referencia')),
    tags        text[] NOT NULL DEFAULT '{}',
    version     integer NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_by  uuid NOT NULL REFERENCES users(id),
    updated_by  uuid NOT NULL REFERENCES users(id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE UNIQUE INDEX documents_title_lower_uq
    ON documents (lower(title))
    WHERE deleted_at IS NULL;
CREATE INDEX documents_tags_gin ON documents USING gin (tags);
CREATE INDEX documents_category_idx ON documents (category) WHERE deleted_at IS NULL;
CREATE INDEX documents_updated_at_idx ON documents (updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX documents_created_by_idx ON documents (created_by);
CREATE INDEX documents_updated_by_idx ON documents (updated_by);

CREATE TABLE document_versions (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id uuid NOT NULL REFERENCES documents(id),
    version     integer NOT NULL CHECK (version >= 1),
    title       text NOT NULL,
    content_md  text NOT NULL,
    category    text NOT NULL,
    tags        text[] NOT NULL DEFAULT '{}',
    edited_by   uuid NOT NULL REFERENCES users(id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (document_id, version)
);

CREATE INDEX document_versions_edited_by_idx ON document_versions (edited_by);

-- +goose Down
DROP TABLE document_versions;
DROP TABLE documents;
