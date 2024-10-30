# Domain Requester for Porkbun

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
