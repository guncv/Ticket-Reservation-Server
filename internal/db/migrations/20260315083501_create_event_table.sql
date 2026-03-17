-- migrate:up
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    title TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC NOT NULL,
    total_tickets INT NOT NULL,
    available_tickets INT NOT NULL DEFAULT 0 CHECK (available_tickets >= 0 AND available_tickets <= total_tickets),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    event_id UUID REFERENCES events(id) ON DELETE RESTRICT NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE RESTRICT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Add quantity column if table was created by an older version of this migration
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS quantity INT NOT NULL DEFAULT 1;

-- migrate:down
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS events;

