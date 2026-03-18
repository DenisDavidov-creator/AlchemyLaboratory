CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE ingredients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    quantity INT NOT NULL DEFAULT 0,
    CONSTRAINT quantity_non_negative CHECK (quantity >= 0)
);


CREATE TABLE recipes (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    brewing_time_seconds INT NOT NULL,

    CONSTRAINT brewing_time_positive CHECK (brewing_time_seconds > 0)
);

CREATE TABLE recipe_ingredients (
     
    recipe_id INT NOT NULL,
    ingredient_id INT NOT NULL,
    quantity_needed INT NOT NULL,
    PRIMARY KEY (recipe_id, ingredient_id),

    CONSTRAINT fk_ingredient FOREIGN KEY(ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE,
    CONSTRAINT fk_recipe FOREIGN KEY(recipe_id) REFERENCES recipes(id) ON DELETE CASCADE,

    CONSTRAINT quantity_needed_positive CHECK (quantity_needed > 0)
);

CREATE TABLE brewing_jobs (
    id SERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    recipe_id INT NOT NULL,
    "status" VARCHAR(20) NOT NULL,
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() ,
    completed_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_recipe FOREIGN KEY(recipe_id) REFERENCES recipes(id) ON DELETE RESTRICT 
);