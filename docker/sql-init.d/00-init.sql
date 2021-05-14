CREATE DATABASE IF NOT EXISTS things;
CREATE USER mainflux WITH LOGIN PASSWORD 'mainflux';
GRANT ALL ON things to mainflux;
