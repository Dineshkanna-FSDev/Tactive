 -- 1. Enable UUID generation so we can use secure, unguessable IDs
 CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    
    -- 2. Users Table: Stores the people making the bookings
    CREATE TABLE users (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMPTZ DEFAULT NOW()
    );
    
    -- 3. Turfs Table: Stores the physical sports grounds
    CREATE TABLE turfs (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        name VARCHAR(255) NOT NULL,
        location VARCHAR(255) NOT NULL,
        price_per_hour NUMERIC NOT NULL CHECK (price_per_hour >= 0)
    );
    
    -- 4. Slots Table: Represents a specific time block for a specific turf
    CREATE TABLE slots (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        turf_id UUID REFERENCES turfs(id) ON DELETE CASCADE,
        start_time TIMESTAMPTZ NOT NULL,
        end_time TIMESTAMPTZ NOT NULL,
        CONSTRAINT start_before_end CHECK (start_time < end_time)
    );
    
    -- 5. Bookings Table: Links a User to a Slot
    CREATE TABLE bookings (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        user_id UUID REFERENCES users(id) ON DELETE CASCADE,
        slot_id UUID REFERENCES slots(id) ON DELETE CASCADE,
        status VARCHAR(50) NOT NULL DEFAULT 'CONFIRMED' CHECK (status IN ('CONFIRMED', 'CANCELLED')),
        created_at TIMESTAMPTZ DEFAULT NOW(),
        cancelled_at TIMESTAMPTZ
    );
    
    -- 6. *** THE MOST IMPORTANT PART ***
    -- This unique index tells PostgreSQL: "For any specific slot_id, you are only allowed to have ONE row where the status is 'CONFIRMED'."
    -- If two users click "Book" at the exact same millisecond, the database will accept the first one and reject the second one with a Conflict error.
    CREATE UNIQUE INDEX unique_active_booking_per_slot 
    ON bookings(slot_id) 
    WHERE status = 'CONFIRMED';