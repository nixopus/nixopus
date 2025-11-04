ALTER TABLE extensions
    DROP CONSTRAINT IF EXISTS extensions_description_check;

ALTER TABLE extensions
    ADD CONSTRAINT extensions_description_check 
    CHECK (LENGTH(description) BETWEEN 10 AND 2000);

