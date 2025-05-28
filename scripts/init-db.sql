-- PostgreSQL Database Schema for GEAV Site
-- Based on the existing JSON data structure

-- Enable UUID extension for generating unique IDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Sequences for auto-incrementing IDs
CREATE SEQUENCE lugares_id_seq START 1;
CREATE SEQUENCE cancoes_id_seq START 1;

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('read', 'write')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on username for faster login queries
CREATE INDEX idx_users_username ON users(username);

-- Tags for lugares
CREATE TABLE tags_lugares (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on tag name for faster lookups
CREATE INDEX idx_tags_lugares_name ON tags_lugares(name);

-- Tags for cancoes
CREATE TABLE tags_cancoes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on tag name for faster lookups
CREATE INDEX idx_tags_cancoes_name ON tags_cancoes(name);

-- Ramos table (shared between lugares and cancoes)
CREATE TABLE ramos (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on ramo name for faster lookups
CREATE INDEX idx_ramos_name ON ramos(name);

-- Lugares table
CREATE TABLE lugares (
    id INTEGER PRIMARY KEY DEFAULT nextval('lugares_id_seq'),
    nome_local VARCHAR(100) NOT NULL,
    nome_dono_local VARCHAR(100),
    telefone_para_contato BIGINT,
    link_google_maps TEXT,
    link_site TEXT,
    endereco_completo TEXT,
    local_publico BOOLEAN NOT NULL DEFAULT false,
    valor_fixo DECIMAL(10, 2) NOT NULL DEFAULT 0,
    valor_individual DECIMAL(10, 2) NOT NULL DEFAULT 0,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for common search fields
CREATE INDEX idx_lugares_nome_local ON lugares(nome_local);
CREATE INDEX idx_lugares_local_publico ON lugares(local_publico);
CREATE INDEX idx_lugares_valor_fixo ON lugares(valor_fixo);
CREATE INDEX idx_lugares_valor_individual ON lugares(valor_individual);

-- Lugares images table (one-to-many relationship)
CREATE TABLE lugares_images (
    id SERIAL PRIMARY KEY,
    lugar_id INTEGER NOT NULL REFERENCES lugares(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on lugar_id for faster lookups
CREATE INDEX idx_lugares_images_lugar_id ON lugares_images(lugar_id);

-- Junction table for lugares and tags (many-to-many)
CREATE TABLE lugares_tags (
    lugar_id INTEGER NOT NULL REFERENCES lugares(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags_lugares(id) ON DELETE CASCADE,
    PRIMARY KEY (lugar_id, tag_id)
);

-- Create indexes for faster lookups
CREATE INDEX idx_lugares_tags_lugar_id ON lugares_tags(lugar_id);
CREATE INDEX idx_lugares_tags_tag_id ON lugares_tags(tag_id);

-- Junction table for lugares and ramos (many-to-many)
CREATE TABLE lugares_ramos (
    lugar_id INTEGER NOT NULL REFERENCES lugares(id) ON DELETE CASCADE,
    ramo_id INTEGER NOT NULL REFERENCES ramos(id) ON DELETE CASCADE,
    PRIMARY KEY (lugar_id, ramo_id)
);

-- Create indexes for faster lookups
CREATE INDEX idx_lugares_ramos_lugar_id ON lugares_ramos(lugar_id);
CREATE INDEX idx_lugares_ramos_ramo_id ON lugares_ramos(ramo_id);

-- Lugares ratings table
CREATE TABLE lugares_ratings (
    id SERIAL PRIMARY KEY,
    lugar_id INTEGER NOT NULL REFERENCES lugares(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (lugar_id, user_id)
);

-- Create indexes for faster lookups and aggregations
CREATE INDEX idx_lugares_ratings_lugar_id ON lugares_ratings(lugar_id);
CREATE INDEX idx_lugares_ratings_user_id ON lugares_ratings(user_id);
CREATE INDEX idx_lugares_ratings_rating ON lugares_ratings(rating);

-- Cancoes table
CREATE TABLE cancoes (
    id INTEGER PRIMARY KEY DEFAULT nextval('cancoes_id_seq'),
    nome VARCHAR(100) NOT NULL,
    link_youtube TEXT,
    letra TEXT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index for common search field
CREATE INDEX idx_cancoes_nome ON cancoes(nome);
CREATE INDEX idx_cancoes_letra ON cancoes USING gin(to_tsvector('portuguese', letra));

-- Junction table for cancoes and tags (many-to-many)
CREATE TABLE cancoes_tags (
    cancao_id INTEGER NOT NULL REFERENCES cancoes(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags_cancoes(id) ON DELETE CASCADE,
    PRIMARY KEY (cancao_id, tag_id)
);

-- Create indexes for faster lookups
CREATE INDEX idx_cancoes_tags_cancao_id ON cancoes_tags(cancao_id);
CREATE INDEX idx_cancoes_tags_tag_id ON cancoes_tags(tag_id);

-- Junction table for cancoes and ramos (many-to-many)
CREATE TABLE cancoes_ramos (
    cancao_id INTEGER NOT NULL REFERENCES cancoes(id) ON DELETE CASCADE,
    ramo_id INTEGER NOT NULL REFERENCES ramos(id) ON DELETE CASCADE,
    PRIMARY KEY (cancao_id, ramo_id)
);

-- Create indexes for faster lookups
CREATE INDEX idx_cancoes_ramos_cancao_id ON cancoes_ramos(cancao_id);
CREATE INDEX idx_cancoes_ramos_ramo_id ON cancoes_ramos(ramo_id);

-- Create materialized view for lugares with average ratings for faster retrieval
CREATE MATERIALIZED VIEW lugares_with_ratings AS
SELECT 
    l.*,
    COALESCE(AVG(lr.rating), 0) AS average_rating,
    COUNT(lr.id) AS rating_count
FROM 
    lugares l
LEFT JOIN 
    lugares_ratings lr ON l.id = lr.lugar_id
GROUP BY 
    l.id;

-- Create index on the materialized view
CREATE INDEX idx_lugares_with_ratings_id ON lugares_with_ratings(id);
CREATE INDEX idx_lugares_with_ratings_average_rating ON lugares_with_ratings(average_rating);
CREATE INDEX idx_lugares_with_ratings_rating_count ON lugares_with_ratings(rating_count);

-- Function to refresh the materialized view
CREATE OR REPLACE FUNCTION refresh_lugares_with_ratings()
RETURNS TRIGGER AS $$
BEGIN
    REFRESH MATERIALIZED VIEW lugares_with_ratings;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to refresh the materialized view when ratings are changed
CREATE TRIGGER refresh_lugares_ratings_view
AFTER INSERT OR UPDATE OR DELETE ON lugares_ratings
FOR EACH STATEMENT
EXECUTE FUNCTION refresh_lugares_with_ratings();

-- Trigger to refresh the materialized view when lugares are changed
CREATE TRIGGER refresh_lugares_view
AFTER INSERT OR UPDATE OR DELETE ON lugares
FOR EACH STATEMENT
EXECUTE FUNCTION refresh_lugares_with_ratings();

-- Initial data for ramos
INSERT INTO ramos (name) VALUES 
('filhotes'),
('lobinho'),
('escoteiro'),
('senior'),
('clã');

-- Initial data for tags_lugares
INSERT INTO tags_lugares (name) VALUES 
('rio'),
('lago'),
('cachoeira'),
('permite_fogueira'),
('mata_fechada'),
('bosque'),
('bambu'),
('rapel'),
('gramado'),
('base_apoio'),
('água_potável'),
('banheiros'),
('trilha'),
('lenha_disponível');

-- Initial data for tags_cancoes
INSERT INTO tags_cancoes (name) VALUES 
('hino'),
('viagem'),
('cerimonia'),
('história'),
('repetição'),
('coreografia'),
('engraçada'),
('longa'),
('animada'),
('reflexiva');

-- Initial admin user
INSERT INTO users (username, password, role) VALUES 
('admin', 'adm_123', 'write'),
('user', 'usr', 'read');

-- Create API logs table
CREATE TABLE api_logs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    service_name TEXT NOT NULL,
    request_id TEXT,
    user_id INTEGER,
    action TEXT,
    resource TEXT,
    resource_id TEXT,
    metadata JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for API logs
CREATE INDEX idx_api_logs_timestamp ON api_logs(timestamp);
CREATE INDEX idx_api_logs_level ON api_logs(level);
CREATE INDEX idx_api_logs_action ON api_logs(action);
CREATE INDEX idx_api_logs_resource ON api_logs(resource);
CREATE INDEX idx_api_logs_user_id ON api_logs(user_id);

-- Comment on tables and columns for documentation
COMMENT ON TABLE users IS 'Users who can access the system';
COMMENT ON TABLE lugares IS 'Places for activities';
COMMENT ON TABLE cancoes IS 'Songs for activities';
COMMENT ON TABLE lugares_ratings IS 'Ratings given by users to places';
COMMENT ON TABLE tags_lugares IS 'Tags that can be applied to places';
COMMENT ON TABLE tags_cancoes IS 'Tags that can be applied to songs';
COMMENT ON TABLE ramos IS 'Scout branches/sections';
COMMENT ON TABLE lugares_images IS 'Images associated with places';
COMMENT ON TABLE lugares_tags IS 'Junction table linking places to tags';
COMMENT ON TABLE lugares_ramos IS 'Junction table linking places to scout branches';
COMMENT ON TABLE cancoes_tags IS 'Junction table linking songs to tags';
COMMENT ON TABLE cancoes_ramos IS 'Junction table linking songs to scout branches';
COMMENT ON MATERIALIZED VIEW lugares_with_ratings IS 'Materialized view of places with their average ratings for faster retrieval';
COMMENT ON TABLE api_logs IS 'Logs of API actions for auditing and monitoring';