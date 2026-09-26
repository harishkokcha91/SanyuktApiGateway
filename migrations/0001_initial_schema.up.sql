-- 0001_initial_schema.up.sql
-- Baseline schema for SanyuktNamdev

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    password TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    age INTEGER NOT NULL,
    image TEXT,
    phone_numbers TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('active', 'inactive', 'pending')),
    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Profiles table (matrimonial profiles)
CREATE TABLE IF NOT EXISTS profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_for TEXT NOT NULL,
    name TEXT NOT NULL,
    image TEXT,
    date_of_birth DATE NOT NULL,
    birth_place TEXT,
    height TEXT,
    complexion TEXT,
    gotra_self TEXT,
    gotra_mother TEXT,
    gotra_grand_mother TEXT,
    manglik BOOLEAN NOT NULL DEFAULT FALSE,
    father_name TEXT,
    father_occupation TEXT,
    mother_name TEXT,
    mother_occupation TEXT,
    siblings TEXT,
    qualification TEXT,
    occupation TEXT,
    annual_income TEXT,
    marital_status TEXT NOT NULL CHECK (marital_status IN ('Single', 'Married', 'Divorced', 'Widowed')),
    address TEXT,
    current_location TEXT,
    status TEXT NOT NULL CHECK (status IN ('pending', 'active', 'inactive')),
    phone_numbers TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);
CREATE INDEX idx_profiles_status ON profiles(status);

-- Events table
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    event_date DATE NOT NULL,
    venue TEXT NOT NULL,
    address TEXT,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    zip_code TEXT,
    country TEXT NOT NULL,
    organizer TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    category TEXT NOT NULL,
    capacity TEXT,
    attendees_registered TEXT,
    status TEXT NOT NULL CHECK (status IN ('upcoming', 'ongoing', 'completed', 'cancelled')),
    image TEXT,
    registration_link TEXT,
    is_online BOOLEAN NOT NULL DEFAULT FALSE,
    ticket_price TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_event_date ON events(event_date);

-- Businesses table
CREATE TABLE IF NOT EXISTS businesses (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT,
    owner TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    whatsapp TEXT,
    location TEXT NOT NULL,
    address TEXT,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    zip_code TEXT,
    country TEXT NOT NULL,
    website TEXT,
    image TEXT,
    status TEXT NOT NULL DEFAULT 'Active' CHECK (status IN ('Active', 'Inactive', 'Pending')),
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    opening_hours TEXT,
    home_delivery BOOLEAN NOT NULL DEFAULT FALSE,
    payment_methods TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_businesses_category ON businesses(category);
CREATE INDEX idx_businesses_status ON businesses(status);
CREATE INDEX idx_businesses_deleted_at ON businesses(deleted_at);

-- Achievements table
CREATE TABLE IF NOT EXISTS achievements (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    achievement_type TEXT NOT NULL,
    achievement TEXT NOT NULL,
    description TEXT,
    date_of_achievement DATE NOT NULL,
    image TEXT,
    status TEXT NOT NULL DEFAULT 'Pending' CHECK (status IN ('Pending', 'Approved', 'Rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_achievements_status ON achievements(status);
CREATE INDEX idx_achievements_date ON achievements(date_of_achievement);