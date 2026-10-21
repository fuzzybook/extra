-- 5_promo_content_broadcaster.up.sql
-- I contenuti delle promo ora indicano un broadcaster (broadcaster_short_id)
-- invece di immagine e link: rimuove dai contenuti esistenti i campi non più
-- usati (url, image_url, mobile_image_url, new_window). Le promo esistenti
-- vanno riassegnate a un broadcaster dal pannello. L'ordine di item e content
-- viene mantenuto.
UPDATE promo
SET item = (
    SELECT COALESCE(jsonb_agg(
        CASE
            WHEN jsonb_typeof(it.value -> 'content') = 'array' THEN jsonb_set(it.value, '{content}', (
                SELECT COALESCE(
                    jsonb_agg(c.value - 'url' - 'image_url' - 'mobile_image_url' - 'new_window' ORDER BY c.ord),
                    '[]'::jsonb
                )
                FROM jsonb_array_elements(it.value -> 'content') WITH ORDINALITY AS c(value, ord)
            ))
            ELSE it.value
        END
        ORDER BY it.ord
    ), '[]'::jsonb)
    FROM jsonb_array_elements(item) WITH ORDINALITY AS it(value, ord)
)
WHERE jsonb_typeof(item) = 'array';
