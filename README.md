# Tactive - Turf Booking & Team Split Platform

Tactive is a full-stack platform for finding, booking, and splitting payments for sports turfs. It allows users to browse nearby turfs, book individual slots, and uniquely create "Team Sessions" to seamlessly split the bill with friends.

## Quick Start (How to Clone and Run)

### Prerequisites
- Node.js (v18+)
- Go (1.20+)
- PostgreSQL (14+)

### Quick Start with Docker (Recommended for Production)

The fastest and most robust way to run Tactive is via Docker.

```bash
git clone <your-repo-url>
cd Tactive
docker-compose up --build -d
```
- Frontend will be available at: `http://localhost` (or `http://localhost:5173`)
- Backend API will be available at: `http://localhost:8080`

*Note: You will still need to execute the initial SQL schema (`backend/migrations/001_initial_schema.sql`) inside your Postgres container and optionally run the seed script!*

---

### Manual Setup (For Development)
The backend API will start on `http://localhost:8080`.

### 4. Start the Frontend (React + Vite)
Open a new terminal window:
```bash
cd frontend
npm install
npm run dev
```
The frontend will start on `http://localhost:5173`. 

---

## Detailed Documentation

### Architecture & Tech Stack

**Frontend:**
- **React.js**: Core UI framework
- **Vite**: Ultra-fast build tool and dev server
- **Vanilla CSS**: Used for UI styling, providing a custom and responsive layout
- **Playwright**: End-to-End (E2E) testing framework

**Backend:**
- **Go (Golang)**: Core API server
- **go-chi**: Lightweight, idiomatic, and composable router for building Go HTTP services
- **PostgreSQL**: Primary relational database
- **pgx**: Pure Go driver and toolkit for PostgreSQL, handling connection pooling
- **JWT**: JSON Web Tokens for stateless user authentication and route protection

### Core Features

#### 1. Real-Time Slot Booking
Users can select a specific date, view available time slots for a specific turf, and seamlessly book them. The database enforces strict consistency using a PostgreSQL `UNIQUE INDEX` to prevent race conditions and double-bookings.

#### 2. Team Split (Pending Sessions)
Instead of paying the full price upfront, users can select a slot and click **TEAM SPLIT**. This generates a shareable unique link that allows friends to pitch in. 
- The target slot is placed in a `PENDING` state and temporarily locked for 15 minutes.
- Uses Advanced PostgreSQL CTEs (`WITH locked_slot AS (...) FOR UPDATE`) to guarantee atomic creation and row-level locking.

#### 3. Turf Exploration
Users can search, filter (Football, Cricket, Tennis, etc.), and view specific details, location, ratings, and pricing for each turf.

### Project Structure

```text
Tactive/
├── backend/
│   ├── cmd/server/
│   │   ├── main.go               # Main Go server entry point
│   │   ├── api_test.go           # Backend API unit tests
│   ├── migrations/
│   │   └── 001_initial_schema.sql # Postgres DB Schema definitions
│   ├── scripts/
│   │   └── seed_slots.go         # Script to generate turfs & time slots
│   └── go.mod                    # Go module dependencies
├── frontend/
│   ├── src/
│   │   ├── App.jsx               # Main React Application & Routes
│   │   ├── index.css             # Main styling stylesheet
│   │   └── main.jsx              # React DOM mounting point
│   ├── tests/e2e/
│   │   └── team_booking.spec.js  # Playwright automated browser tests
│   ├── package.json              # NPM dependencies
│   └── vite.config.js            # Vite configuration & proxy settings
```

### End-to-End Testing (Playwright)

Tactive utilizes Playwright to ensure critical user journeys (like logging in and booking a team session) function seamlessly. 

To run the automated tests:
```bash
cd frontend
npm install @playwright/test

# Download required browsers (first-time only)
npx playwright install --with-deps chromium

# Run the tests
npx playwright test
```

### API References

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/auth/login` | Authenticate and retrieve a JWT |
| `GET` | `/api/turfs` | Fetch a list of all available sports turfs |
| `GET` | `/api/slots` | Fetch all slots and their booking status |
| `POST` | `/api/bookings` | Book a specific time slot (Requires JWT) |
| `POST` | `/api/bookings/{id}/cancel` | Cancel an existing booking (Requires JWT) |
| `POST` | `/api/team-sessions` | Create a pending team session lock (Requires JWT) |
| `GET` | `/api/user/bookings` | Fetch past/upcoming bookings for the logged-in user |
