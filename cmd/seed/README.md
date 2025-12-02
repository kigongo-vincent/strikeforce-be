# Database Seeder CLI Tool

A comprehensive CLI tool for seeding the database with realistic Ugandan data for testing purposes. This tool generates large-scale test data including organizations, users, projects, applications, and more.

## Quick Start

```bash
# Build the seeder
cd backend
go build -o bin/seed ./cmd/seed

# Run with default values (10 orgs, 100 users, 50 projects, 200 applications)
./bin/seed

# Run with custom values for large-scale testing
./bin/seed [organizations] [users] [projects] [applications]

# Example: Create 20 organizations, 500 users, 200 projects, 1000 applications
./bin/seed 20 500 200 1000

# Example: Create massive dataset for performance testing
./bin/seed 50 2000 500 5000
```

## Features

- **Realistic Ugandan Data**: 
  - 20+ real Ugandan universities (Makerere, Kyambogo, UCU, etc.)
  - 40+ real Ugandan companies (MTN, Airtel, banks, NGOs, etc.)
  - 200+ common first and last names
  - 30+ academic departments
  - 30+ academic courses
  - 30+ project titles and descriptions
  - 50+ technical skills

- **Comprehensive Seeding**: Creates complete data relationships:
  - Organizations (universities & partners)
  - Departments (3-8 per university)
  - Courses (2-5 per department)
  - Users (students, supervisors, partners, university-admins)
  - Students (linked to courses)
  - Supervisors (linked to departments)
  - Projects (with budgets, deadlines, skills, statuses)
  - Applications (student applications to projects)
  - Milestones (2-5 per project)
  - Portfolio Items (from completed milestones)

- **Password**: All seeded users have password `1234567890`

- **Relationships**: Properly links all entities maintaining referential integrity

- **Large Scale**: Can generate thousands of records for comprehensive testing

## Data Generated

### Organizations
- **Universities**: Makerere University, Kyambogo University, UCU, etc.
- **Partners**: MTN Uganda, Airtel, banks, NGOs, tech companies, etc.
- Each organization gets an admin user (university-admin or partner role)

### Departments & Courses
- Departments: Computer Science, Engineering, Business, Medicine, etc.
- Courses: Bachelor programs matching departments
- Automatically linked to universities

### Users
- **Students**: Linked to courses and universities
- **Supervisors**: Linked to departments
- **Partners**: Linked to partner organizations
- **University Admins**: One per university
- All with realistic names, emails, phone numbers, and skills

### Projects
- Created by partner users
- Linked to departments
- Random budgets (USD or UGX)
- Various statuses (pending, in-progress, completed, on-hold)
- Skills arrays
- Deadlines (30-180 days from now)

### Applications
- Students applying to projects
- Various statuses (SUBMITTED, SHORTLISTED, ACCEPTED, etc.)
- Individual applications

### Milestones
- 2-5 milestones per project
- Various statuses
- Budget amounts
- Due dates

### Portfolio Items
- Auto-generated from completed milestones
- Linked to students who worked on projects
- Includes ratings, complexity, amounts, on-time status

## Data Details

- **Email Format**: `firstname.lastname.{index}@example.com`
- **Phone Format**: `+2567XXXXXXXX` (Ugandan mobile format)
- **Location**: All users located in "Kampala, Uganda"
- **Skills**: 3-8 random technical skills per user
- **Passwords**: All users: `1234567890`

## Example Usage

```bash
# Small dataset for quick testing
./bin/seed 5 50 20 100

# Medium dataset for feature testing
./bin/seed 10 200 100 500

# Large dataset for performance testing
./bin/seed 30 1000 500 3000

# Massive dataset for stress testing
./bin/seed 50 5000 2000 10000
```

## Notes

- The seeder maintains referential integrity
- All relationships are properly established
- Data is realistic and contextually appropriate for Uganda
- Can be run multiple times (will create duplicate data)
- For clean seeding, truncate tables first if needed

