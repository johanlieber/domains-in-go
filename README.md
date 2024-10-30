# Domain Requester for Porkbun

API Setup
---------
add auth to the routes

DB Setup
--------
model domain {
  id String @id @default(cuid())
  identifier String
  requestType String
  baseDomain String
  targetHost String
  TTL Int
  description String
  createdAt DateTime @default(now())
}
