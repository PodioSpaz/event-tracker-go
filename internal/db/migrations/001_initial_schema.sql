-- Event Tracker Initial Schema
-- Created: 2025-12-12

-- Activities table (Events for MVP, Gatherings supported in schema)
CREATE TABLE IF NOT EXISTS activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    date DATE NOT NULL,
    location TEXT NOT NULL,
    activity_type TEXT NOT NULL DEFAULT 'event' CHECK(activity_type IN ('event', 'gathering')),
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'cancelled', 'completed')),
    requires_registration BOOLEAN NOT NULL DEFAULT 1,
    is_free BOOLEAN NOT NULL DEFAULT 0,
    fee TEXT DEFAULT '0.00',  -- Store as TEXT for decimal precision
    max_capacity INTEGER,
    estimated_head_count INTEGER,
    actual_head_count INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- People table
CREATE TABLE IF NOT EXISTS people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT DEFAULT '',
    phone TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (email != '' OR phone != '')
);

-- Attendees table (junction table linking activities and people)
CREATE TABLE IF NOT EXISTS attendees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    activity_id INTEGER NOT NULL,
    person_id INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'participant' CHECK(role IN ('participant', 'volunteer', 'worship_team', 'workshop_leader')),
    payment_status TEXT NOT NULL DEFAULT 'unpaid' CHECK(payment_status IN ('paid', 'unpaid', 'waived')),
    payment_amount TEXT DEFAULT '0.00',  -- Store as TEXT for decimal precision
    payment_date DATE,
    registration_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT DEFAULT '',
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (person_id) REFERENCES people(id) ON DELETE CASCADE,
    UNIQUE(activity_id, person_id)
);

-- Expenditures table (deferred to v0.2.0, but schema included for future use)
CREATE TABLE IF NOT EXISTS expenditures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    activity_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    amount TEXT NOT NULL,  -- Store as TEXT for decimal precision
    category TEXT NOT NULL CHECK(category IN ('venue', 'food', 'supplies', 'transport', 'other')),
    date DATE NOT NULL,
    receipt_path TEXT,
    notes TEXT DEFAULT '',
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT 1
);

-- Seed default roles
INSERT OR IGNORE INTO roles (name, display_name, description) VALUES
    ('participant', 'Participant', 'Regular event participant'),
    ('volunteer', 'Volunteer', 'Event volunteer helper'),
    ('worship_team', 'Worship Team', 'Worship team member'),
    ('workshop_leader', 'Workshop Leader', 'Leads workshops during the event');

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_activities_date ON activities(date);
CREATE INDEX IF NOT EXISTS idx_activities_status ON activities(status);
CREATE INDEX IF NOT EXISTS idx_activities_type ON activities(activity_type);

CREATE INDEX IF NOT EXISTS idx_people_email ON people(email);
CREATE INDEX IF NOT EXISTS idx_people_last_name ON people(last_name);

CREATE INDEX IF NOT EXISTS idx_attendees_activity ON attendees(activity_id);
CREATE INDEX IF NOT EXISTS idx_attendees_person ON attendees(person_id);
CREATE INDEX IF NOT EXISTS idx_attendees_payment_status ON attendees(payment_status);

CREATE INDEX IF NOT EXISTS idx_expenditures_activity ON expenditures(activity_id);
CREATE INDEX IF NOT EXISTS idx_expenditures_category ON expenditures(category);

-- Create triggers for updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_activities_timestamp
    AFTER UPDATE ON activities
BEGIN
    UPDATE activities SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_people_timestamp
    AFTER UPDATE ON people
BEGIN
    UPDATE people SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
