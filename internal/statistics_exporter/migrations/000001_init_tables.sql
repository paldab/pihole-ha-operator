-- +goose Up
CREATE TABLE pihole_database_checkpoint (
    cluster_uuid TEXT NOT NULL,
    source_uuid TEXT NOT NULL,

    last_exported_id BIGINT NOT NULL,

    PRIMARY KEY (
        cluster_uuid,
        source_uuid
    )
);


CREATE TABLE pihole_queries (
    cluster_uuid TEXT NOT NULL,
    source_uuid TEXT NOT NULL,

    local_query_id BIGINT NOT NULL,

    query_time DOUBLE PRECISION NOT NULL,

    
    type TEXT NOT NULL,
    status TEXT NOT NULL,

    client TEXT NOT NULL,
    domain TEXT NOT NULL,

    reply_time DOUBLE PRECISION,

    PRIMARY KEY (
        cluster_uuid,
        source_uuid,
        local_query_id
    )
);

-- +goose Down
DROP TABLE pihole_database_checkpoint
DROP TABLE pihole_queries

