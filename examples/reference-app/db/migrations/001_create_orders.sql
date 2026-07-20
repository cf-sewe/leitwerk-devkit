-- 001_create_orders — the schema the reference-app's orders ride on.
--
-- Under Leitwerk's default tiers, anything in db/migrations/ (and any *.sql) is
-- T2: an irreversible / data path, the worst blast radius. Editing this file
-- escalates the gate to the full check set (T2 = lint types tests drift sast
-- erosion). It exists to demonstrate that escalation on real code.
CREATE TABLE orders (
    id          INTEGER PRIMARY KEY,
    total_cents INTEGER NOT NULL CHECK (total_cents >= 0)
);
