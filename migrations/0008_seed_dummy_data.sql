-- Profiles
INSERT INTO profiles (user_id, name, age, location, status)
VALUES (1, 'Amit Sharma', 29, 'Delhi', 'Approved'),
       (2, 'Priya Verma', 26, 'Mumbai', 'Pending');

-- Businesses
INSERT INTO businesses (user_id, name, category, location, status)
VALUES (1, 'Sharma Electronics', 'Retail', 'Delhi', 'Approved'),
       (2, 'Verma Consulting', 'IT Services', 'Bangalore', 'Pending');

-- Events
INSERT INTO events (user_id, title, event_date, location, status)
VALUES (1, 'Tech Meetup', '2026-10-15', 'Delhi', 'Approved'),
       (2, 'Startup Pitch', '2026-11-01', 'Mumbai', 'Pending');

-- Achievements
INSERT INTO achievements (user_id, title, description, status)
VALUES (1, 'Best Innovator Award', 'Awarded by NASSCOM', 'Approved'),
       (2, 'Hackathon Winner', 'Won local hackathon', 'Pending');
