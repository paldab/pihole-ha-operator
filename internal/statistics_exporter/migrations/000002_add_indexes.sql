-- +goose Up
CREATE INDEX idx_query_time
ON pihole_queries(query_time);

CREATE INDEX idx_query_domain
ON pihole_queries(domain);

CREATE INDEX idx_query_client
ON pihole_queries(client);

-- +goose Down
DROP INDEX idx_query_time 
ON pihole_queries(query_time);

DROP INDEX idx_query_domain
ON pihole_queries(domain);

DROP INDEX idx_query_client
ON pihole_queries(client);

