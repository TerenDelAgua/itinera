-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Trips
CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    base_currency CHAR(3) DEFAULT 'EUR',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Expenses
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID REFERENCES trips(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    amount NUMERIC(12,2) NOT NULL,
    currency CHAR(3) NOT NULL,
    converted_amount NUMERIC(12,2),
    category VARCHAR(50) DEFAULT 'other',
    date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Places (destinos dentro de un viaje: Tokyo, Nara, Kyoto...)
CREATE TABLE places (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID REFERENCES trips(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    lat DECIMAL(10, 8),      -- Para mapas
    lng DECIMAL(11, 8),
    start_date DATE,         -- Opcional: si el lugar tiene fechas específicas
    end_date DATE,
    order_index INTEGER DEFAULT 0, -- Para ordenar lugares en el itinerario
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Activities (Agenda)
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    place_id UUID REFERENCES places(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    scheduled_at TIMESTAMPTZ NOT NULL,
    type VARCHAR(50) DEFAULT 'visit',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes para rendimiento
CREATE INDEX idx_expenses_trip_id ON expenses(trip_id);
CREATE INDEX idx_places_trip_id ON places(trip_id);
CREATE INDEX idx_activities_place_id ON activities(place_id);
CREATE INDEX idx_activities_scheduled ON activities(scheduled_at);