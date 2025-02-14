CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    author VARCHAR(255),
    genres TEXT[],   
    release_year INT,
    number_of_pages INT,
    image_url VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

