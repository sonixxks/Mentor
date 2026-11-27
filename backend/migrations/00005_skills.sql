-- +goose Up
-- Единый словарь навыков: анкеты, программы и фильтры каталога ссылаются на name.
CREATE TABLE skills (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    is_popular BOOLEAN      NOT NULL DEFAULT FALSE
);

-- Названия сравниваем без учёта регистра: «go» и «Go» — один навык.
CREATE UNIQUE INDEX uq_skills_name_lower ON skills (lower(name));

INSERT INTO skills (name, is_popular) VALUES
    ('Python', TRUE),
    ('JavaScript', TRUE),
    ('TypeScript', TRUE),
    ('Go', TRUE),
    ('Java', TRUE),
    ('C#', FALSE),
    ('C++', FALSE),
    ('Kotlin', FALSE),
    ('Swift', FALSE),
    ('PHP', FALSE),
    ('React', TRUE),
    ('Vue', FALSE),
    ('Angular', FALSE),
    ('HTML/CSS', TRUE),
    ('Node.js', FALSE),
    ('Django', FALSE),
    ('FastAPI', FALSE),
    ('Spring', FALSE),
    ('SQL', TRUE),
    ('PostgreSQL', FALSE),
    ('MongoDB', FALSE),
    ('Redis', FALSE),
    ('Docker', TRUE),
    ('Kubernetes', FALSE),
    ('Linux', FALSE),
    ('Git', TRUE),
    ('CI/CD', FALSE),
    ('Алгоритмы и структуры данных', TRUE),
    ('Архитектура ПО', FALSE),
    ('Тестирование', FALSE),
    ('Машинное обучение', TRUE),
    ('Анализ данных', TRUE),
    ('Android', FALSE),
    ('iOS', FALSE),
    ('UI/UX дизайн', TRUE),
    ('Figma', FALSE),
    ('Продуктовый менеджмент', FALSE),
    ('Системный анализ', FALSE),
    ('Информационная безопасность', FALSE),
    ('DevOps', FALSE);

-- +goose Down
DROP TABLE IF EXISTS skills;
