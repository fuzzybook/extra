-- Inizializzazione dei database per i microservizi Encore
SELECT 'CREATE DATABASE auth' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth')\gexec
SELECT 'CREATE DATABASE profiles' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'profiles')\gexec
SELECT 'CREATE DATABASE advertising' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'advertising')\gexec
SELECT 'CREATE DATABASE images' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'images')\gexec
SELECT 'CREATE DATABASE registration' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'registration')\gexec
SELECT 'CREATE DATABASE extrahome' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'extrahome')\gexec
SELECT 'CREATE DATABASE broadcasters' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'broadcasters')\gexec
SELECT 'CREATE DATABASE videochat' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'videochat')\gexec
