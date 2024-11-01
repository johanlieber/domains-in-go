# Domain Requester for Porkbun

API Setup
---------
add auth to the routes

DB Setup
--------
	CREATE TABLE owned_domains IF NOT EXISTS (
		status VARCHAR(100) NOT NULL,
		name VARCHAR(100) UNIQUE NOT NULL,
		expires_at TIMESTAMP,
		obtained_at TIMESTAMP,
		created_at DEFAULT CURRENT_TIMESTAMP
		PRIMARY KEY (name)
	)

Things to do
------------
- Take API results and store in DB.
  - Upsert API results
- Pull from DB instead of API? Maybe on first load.
- Dropdown for Dashboard with the results
  - Potentially have Links page connect to dashboard
- Implement the porkbun API part
  - Also make sure optional name field is conditionally added on
