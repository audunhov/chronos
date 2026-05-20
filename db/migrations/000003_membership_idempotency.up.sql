-- Sikre at en bruker bare kan ha ett medlemskap per organisasjon (default policy)
CREATE UNIQUE INDEX IF NOT EXISTS idx_membership_user_org_unique ON membership_view(user_id, org_id);

-- Vi kan senere utvide dette med mer kompleks ltree-basert sjekk 
-- for å håndtere umbrella-eksklusivitet.
