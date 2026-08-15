ALTER TABLE bookings DROP CONSTRAINT bookings_status_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check CHECK (status IN ('CONFIRMED', 'CANCELLED', 'PENDING'));

DROP INDEX unique_active_booking_per_slot;
CREATE UNIQUE INDEX unique_active_booking_per_slot ON bookings (slot_id) WHERE status IN ('CONFIRMED', 'PENDING');

ALTER TABLE bookings ADD COLUMN expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE bookings ADD COLUMN total_amount DECIMAL(10,2) DEFAULT 0;
ALTER TABLE bookings ADD COLUMN paid_amount DECIMAL(10,2) DEFAULT 0;
ALTER TABLE bookings ADD COLUMN is_split_pay BOOLEAN DEFAULT FALSE;

CREATE TABLE team_contributions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID REFERENCES bookings(id) ON DELETE CASCADE,
    contributor_name VARCHAR(100) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
