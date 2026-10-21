-- 7_promo_content_broadcaster_object.up.sql
-- Il riferimento al broadcaster passa da content.broadcaster_short_id a
-- content.broadcaster.short_id (unico campo salvato del broadcaster).
-- L'ordine di item e content viene mantenuto.
UPDATE promo
SET item = (
    SELECT COALESCE(jsonb_agg(
        CASE
            WHEN jsonb_typeof(it.value -> 'content') = 'array' THEN jsonb_set(it.value, '{content}', (
                SELECT COALESCE(jsonb_agg(
                    CASE
                        WHEN coalesce(c.value ->> 'broadcaster_short_id', '') <> ''
                            THEN (c.value - 'broadcaster_short_id' - 'broadcaster')
                                 || jsonb_build_object('broadcaster', jsonb_build_object('short_id', c.value ->> 'broadcaster_short_id'))
                        ELSE c.value - 'broadcaster_short_id'
                    END
                    ORDER BY c.ord
                ), '[]'::jsonb)
                FROM jsonb_array_elements(it.value -> 'content') WITH ORDINALITY AS c(value, ord)
            ))
            ELSE it.value
        END
        ORDER BY it.ord
    ), '[]'::jsonb)
    FROM jsonb_array_elements(item) WITH ORDINALITY AS it(value, ord)
)
WHERE jsonb_typeof(item) = 'array';
