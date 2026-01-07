-- init.sql

-- 1. Users & Auth
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    -- MATCHING YOUR GO CODE HERE:
    password_hash VARCHAR(255) NOT NULL, 
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    student_name VARCHAR(100) NOT NULL,
    guardian_name VARCHAR(100) NOT NULL,
    school_name VARCHAR(100) NOT NULL,
    grade VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 2. System Settings
CREATE TABLE IF NOT EXISTS system_settings (
    setting_key VARCHAR(50) PRIMARY KEY,
    setting_value TEXT
);
-- Default event dates (You can change these later in Admin Config)
INSERT INTO system_settings (setting_key, setting_value) VALUES 
('event_date_1', '2025-08-01'),
('event_date_2', '2025-08-02')
ON CONFLICT DO NOTHING;


-- 3. Classes & Instructors
CREATE TABLE IF NOT EXISTS classes (
    class_id SERIAL PRIMARY KEY,
    class_name TEXT NOT NULL,
    syllabus_pdf_url TEXT,
    room_number VARCHAR(50),
    room_name TEXT,
    registration_start_at TIMESTAMPTZ NOT NULL,
    registration_end_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS instructors (
    instructor_id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS class_instructors (
    class_id INT NOT NULL,
    instructor_id INT NOT NULL,
    PRIMARY KEY (class_id, instructor_id),
    CONSTRAINT fk_class FOREIGN KEY (class_id) REFERENCES classes(class_id) ON DELETE CASCADE,
    CONSTRAINT fk_instructor FOREIGN KEY (instructor_id) REFERENCES instructors(instructor_id) ON DELETE CASCADE
);


-- 4. Sessions
CREATE TABLE IF NOT EXISTS class_sessions (
    session_id SERIAL PRIMARY KEY,
    class_id INT NOT NULL,
    day_sequence INT NOT NULL, -- 1 or 2
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    capacity INT NOT NULL,
    current_enrolled_count INT DEFAULT 0,
    CONSTRAINT fk_session_class FOREIGN KEY (class_id) REFERENCES classes(class_id) ON DELETE CASCADE
);


-- 5. Enrollments
CREATE TABLE IF NOT EXISTS session_enrollments (
    enrollment_id SERIAL PRIMARY KEY,
    session_id INT NOT NULL,
    user_profile_id INT NOT NULL,
    registered_at TIMESTAMPTZ DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'confirmed',
    
    UNIQUE(session_id, user_profile_id),

    CONSTRAINT fk_enrollment_session FOREIGN KEY (session_id) REFERENCES class_sessions(session_id) ON DELETE CASCADE,
    CONSTRAINT fk_enrollment_profile FOREIGN KEY (user_profile_id) REFERENCES user_profiles(id) ON DELETE CASCADE
);
