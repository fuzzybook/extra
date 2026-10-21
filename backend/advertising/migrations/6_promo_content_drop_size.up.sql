-- 6_promo_content_drop_size.up.sql
-- Le dimensioni non sono più del contenuto della promo: arrivano dalla foto
-- del broadcaster. Rimuove width e height dai contenuti salvati, mantenendo
-- l'ordine di item e content.
UPDATE promo
SET item = (
    SELECT COALESCE(jsonb_agg(
        CASE
            WHEN jsonb_typeof(it.value -> 'content') = 'array' THEN jsonb_set(it.value, '{content}', (
                SELECT COALESCE(jsonb_agg(c.value - 'width' - 'height' ORDER BY c.ord), '[]'::jsonb)
                FROM jsonb_array_elements(it.value -> 'content') WITH ORDINALITY AS c(value, ord)
            ))
            ELSE it.value
        END
        ORDER BY it.ord
    ), '[]'::jsonb)
    FROM jsonb_array_elements(item) WITH ORDINALITY AS it(value, ord)
)
WHERE jsonb_typeof(item) = 'array';
