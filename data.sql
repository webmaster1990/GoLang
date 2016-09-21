-- TEST DATA
INSERT INTO organizations (organization_id, organization_name, mission, vision, program, objective)
VALUES ('5827BDC9-1372-46EF-904C-58977202DF63', 'Team Rocket', 'To protect the world from devestation!',
        'To unite all people within our nation!', 'To denounce the evils of truth and love!',
        'To extend our reach to the stars above!');

INSERT INTO users (user_id, organization_id, email, password, full_name, is_admin, is_active)
VALUES ('9A46D26D-F6D3-4924-99C2-5E4FE4242441', '5827BDC9-1372-46EF-904C-58977202DF63', 'giovanni@teamrocket.org',
        '$2b$12$5sjei9djq9j8BjE4NrMPguy078iBmTh2HSzqqE6n9vtPf2wsV4gO6', 'Giovanni', TRUE, TRUE),
  ('864656E0-FC45-4C83-9310-FEE77BE1D5CA', '5827BDC9-1372-46EF-904C-58977202DF63', 'james@teamrocket.org',
   '$2b$12$5sjei9djq9j8BjE4NrMPguy078iBmTh2HSzqqE6n9vtPf2wsV4gO6', 'James', FALSE, TRUE),
  ('C0D3EAF0-8404-42B6-A04F-5546B38BB0A0', '5827BDC9-1372-46EF-904C-58977202DF63', 'jessie@teamrocket.org',
   '$2b$12$5sjei9djq9j8BjE4NrMPguy078iBmTh2HSzqqE6n9vtPf2wsV4gO6', 'Jessie', FALSE, TRUE),
  ('E3071BF8-6986-4A45-A6D0-1A238DC91B47', '5827BDC9-1372-46EF-904C-58977202DF63', 'meowth@teamrocket.org',
   '$2b$12$5sjei9djq9j8BjE4NrMPguy078iBmTh2HSzqqE6n9vtPf2wsV4gO6', 'Meowth', FALSE, TRUE);
