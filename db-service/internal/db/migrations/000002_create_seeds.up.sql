INSERT INTO ingredients (name, description, quantity) VALUES
    ('Dragon Scale',    'Rare scale from a fire dragon',     100),
    ('Moonflower',      'Flower that blooms only at night',  50),
    ('Shadow Essence',  'Dark energy distilled into liquid', 30),
    ('Stardust',        'Collected from falling stars',      75),
    ('Phoenix Feather', 'Feather from a reborn phoenix',     20);

INSERT INTO recipes (name, description, brewing_time_seconds) VALUES
    ('Love Elixir',    'Makes the drinker fall in love',     10),
    ('Invisibility',   'Renders the drinker invisible',      20),
    ('Dragon Breath',  'Grants fire breathing for 1 hour',   25);

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity_needed) VALUES
    (1, 1, 2),  -- Love Elixir needs 2 Dragon Scales
    (1, 2, 3),  -- Love Elixir needs 3 Moonflowers
    (2, 3, 5),  -- Invisibility needs 5 Shadow Essence
    (2, 4, 2),  -- Invisibility needs 2 Stardust
    (3, 1, 3),  -- Dragon Breath needs 3 Dragon Scales
    (3, 5, 1);  -- Dragon Breath needs 1 Phoenix Feather