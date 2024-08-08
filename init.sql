create database gsatlabbest;

use gsatlabbest;

CREATE TABLE users (
    user_id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    salt VARCHAR(32) NOT NULL,
    agree_privacy BOOLEAN NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    school_list TEXT,
    permit INT DEFAULT 1
);

CREATE TABLE system_secrets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL
);

INSERT INTO system_secrets (name, content) VALUES ('smtp_config', 'host:smtp.example.com;port:587;user:your_email@example.com;password:your_password');

CREATE TABLE tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    purpose VARCHAR(255) NOT NULL,
    user_id INT NOT NULL,
    token VARCHAR(1024) NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE 113_year (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tag VARCHAR(255),
    code VARCHAR(255),
    combined_name VARCHAR(255),
    enrollment_quota VARCHAR(255),
    screening_date VARCHAR(255),
    subject JSON,
    subject_criteria JSON,
    full_json JSON,
    FULLTEXT(combined_name, code)
);

CREATE TABLE 112_year (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tag VARCHAR(255),
    code VARCHAR(255),
    combined_name VARCHAR(255),
    enrollment_quota VARCHAR(255),
    screening_date VARCHAR(255),
    subject JSON,
    subject_criteria JSON,
    full_json JSON,
    FULLTEXT(combined_name, code)
);
